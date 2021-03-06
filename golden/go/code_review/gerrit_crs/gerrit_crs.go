// Package gerrit_crs provides a client for Gold's interaction with
// the Gerrit code review system.
package gerrit_crs

import (
	"context"
	"errors"
	"sort"
	"strconv"

	"golang.org/x/time/rate"

	"go.skia.org/infra/go/gerrit"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/vcsinfo"
	"go.skia.org/infra/golden/go/code_review"
)

const (
	// These values are arbitrary guesses, roughly based on the values for gitiles.
	maxQPS   = rate.Limit(4.0)
	maxBurst = 20
)

type CRSImpl struct {
	gClient gerrit.GerritInterface
	rl      *rate.Limiter
}

func New(client gerrit.GerritInterface) *CRSImpl {
	return &CRSImpl{
		gClient: client,
		rl:      rate.NewLimiter(maxQPS, maxBurst),
	}
}

var invalidID = errors.New("invalid id - must be integer")

// GetChangeList implements the code_review.Client interface.
func (c *CRSImpl) GetChangeList(ctx context.Context, id string) (code_review.ChangeList, error) {
	cl, err := c.getGerritCL(ctx, id)
	if err != nil {
		return code_review.ChangeList{}, err
	}
	return code_review.ChangeList{
		SystemID: strconv.FormatInt(cl.Issue, 10),
		Owner:    cl.Owner.Email,
		Status:   statusToEnum(cl.Status),
		Subject:  cl.Subject,
		Updated:  cl.Updated,
	}, nil
}

// statusToEnum converts a gerrit status string into a CLStatus enum.
func statusToEnum(g string) code_review.CLStatus {
	switch g {
	case gerrit.CHANGE_STATUS_NEW:
		return code_review.Open
	case gerrit.CHANGE_STATUS_ABANDONED:
		return code_review.Abandoned
	case gerrit.CHANGE_STATUS_MERGED:
		return code_review.Landed
	}
	return code_review.Open
}

// GetPatchSets implements the code_review.Client interface.
func (c *CRSImpl) GetPatchSets(ctx context.Context, clID string) ([]code_review.PatchSet, error) {
	cl, err := c.getGerritCL(ctx, clID)
	if err != nil {
		return nil, err
	}
	var xps []code_review.PatchSet
	for _, p := range cl.Patchsets {
		xps = append(xps, code_review.PatchSet{
			SystemID:     p.ID,
			ChangeListID: clID,
			Order:        int(p.Number),
			GitHash:      p.ID,
		})
	}
	// Gerrit probably returns them in order, but this ensures it.
	sort.Slice(xps, func(i, j int) bool {
		return xps[i].Order < xps[j].Order
	})

	return xps, nil
}

// GetChangeListIDForCommit implements the code_review.Client interface.
func (c *CRSImpl) GetChangeListIDForCommit(_ context.Context, commit *vcsinfo.LongCommit) (string, error) {
	if commit == nil {
		return "", skerr.Fmt("commit cannot be nil")
	}
	i, err := c.gClient.ExtractIssueFromCommit(commit.Body)
	if err != nil {
		sklog.Debugf("Could not find gerrit issue in %q: %s", commit.Body, err)
		return "", code_review.ErrNotFound
	}
	return strconv.FormatInt(i, 10), nil
}

// CommentOn implements the code_review.Client interface.
func (c *CRSImpl) CommentOn(ctx context.Context, clID, message string) error {
	sklog.Infof("Commenting on Gerrit CL %s with message %q", clID, message)
	cl, err := c.getGerritCL(ctx, clID)
	if err != nil {
		return err
	}
	return skerr.Wrapf(c.gClient.AddComment(ctx, cl, message), "commenting on gerrit CL %s", clID)
}

// System implements the code_review.Client interface.
func (c *CRSImpl) System() string {
	return "gerrit"
}

func (c *CRSImpl) getGerritCL(ctx context.Context, clID string) (*gerrit.ChangeInfo, error) {
	i, err := strconv.ParseInt(clID, 10, 64)
	if err != nil {
		return nil, invalidID
	}
	// Respect the rate limit.
	if err := c.rl.Wait(ctx); err != nil {
		return nil, skerr.Wrap(err)
	}
	cl, err := c.gClient.GetIssueProperties(ctx, i)
	if err == gerrit.ErrNotFound {
		return nil, code_review.ErrNotFound
	}
	if err != nil {
		return nil, skerr.Wrapf(err, "fetching CL from gerrit with id %d", i)
	}
	return cl, nil
}

// Make sure CRSImpl fulfills the code_review.Client interface.
var _ code_review.Client = (*CRSImpl)(nil)
