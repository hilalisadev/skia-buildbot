// Package docker is for running Dockerfiles.
package docker

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.skia.org/infra/go/exec"
	"go.skia.org/infra/go/testutils/unittest"
	"go.skia.org/infra/task_driver/go/td"
)

func TestBuild(t *testing.T) {
	unittest.MediumTest(t)
	// Strip our PATH so we find our version of `docker` which is in the
	// test_bin directory. Then add `/bin` to the PATH since we are running a
	// Bash shell.
	_, filename, _, _ := runtime.Caller(0)
	dockerCmd = filepath.Join(filepath.Dir(filename), "test_bin", "docker_mock")

	type args struct {
		tag string
	}
	tests := []struct {
		name                 string
		args                 args
		subSteps             int
		timeout              time.Duration
		expected             td.StepResult
		expectedFirstSubStep td.StepResult
		wantErr              bool
		buildArgs            map[string]string
	}{
		{
			name: "success",
			args: args{
				tag: "success",
			},
			subSteps:             7,
			timeout:              time.Minute,
			expected:             td.STEP_RESULT_SUCCESS,
			expectedFirstSubStep: td.STEP_RESULT_SUCCESS,
			wantErr:              false,
			buildArgs:            map[string]string{"arg1": "value1"},
		},
		{
			name: "failure",
			args: args{
				tag: "failure",
			},
			subSteps:             0,
			timeout:              time.Minute,
			expected:             td.STEP_RESULT_SUCCESS,
			expectedFirstSubStep: td.STEP_RESULT_FAILURE,
			wantErr:              true,
			buildArgs:            nil,
		},
		{
			name: "failure_no_output",
			args: args{
				tag: "failure_no_output",
			},
			subSteps:             0,
			timeout:              time.Minute,
			expected:             td.STEP_RESULT_SUCCESS,
			expectedFirstSubStep: td.STEP_RESULT_FAILURE,
			wantErr:              true,
			buildArgs:            nil,
		},
		{
			name: "timeout",
			args: args{
				tag: "timeout",
			},
			subSteps:             0,
			timeout:              time.Millisecond,
			expected:             td.STEP_RESULT_SUCCESS,
			expectedFirstSubStep: td.STEP_RESULT_FAILURE,
			wantErr:              true,
			buildArgs:            nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := td.StartTestRun(t)
			defer tr.Cleanup()

			// Root-level step.
			ctx := tr.Root()

			// Add a timeout.
			ctx, cancel := context.WithTimeout(ctx, tt.timeout)
			defer cancel()

			if err := BuildHelper(ctx, ".", tt.args.tag, "test_config_dir", tt.buildArgs); (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Ensure that we got the expected step results.
			results := tr.EndRun(false, nil)

			require.Equal(t, tt.subSteps, len(results.Steps[0].Steps))
			require.Equal(t, tt.expected, results.Result)
			require.Equal(t, tt.expectedFirstSubStep, results.Steps[0].Result)
		})
	}

}

func TestLogin(t *testing.T) {
	unittest.SmallTest(t)

	_ = td.RunTestSteps(t, false, func(ctx context.Context) error {
		mockRun := &exec.CommandCollector{}
		mockRun.SetDelegateRun(func(ctx context.Context, cmd *exec.Command) error {
			require.Equal(t, dockerCmd, cmd.Name)
			require.Equal(t, []string{"--config", "test_config_dir", "login", "-u", "oauth2accesstoken", "--password-stdin", "https://gcr.io"}, cmd.Args)
			require.Equal(t, "", cmd.Dir)
			require.Equal(t, strings.NewReader("token"), cmd.Stdin)
			return nil
		})
		ctx = td.WithExecRunFn(ctx, mockRun.Run)

		err := Login(ctx, "token", "https://gcr.io", "test_config_dir")
		require.NoError(t, err)

		return nil
	})
}

func TestRun(t *testing.T) {
	unittest.SmallTest(t)

	_ = td.RunTestSteps(t, false, func(ctx context.Context) error {
		mockRun := &exec.CommandCollector{}
		mockRun.SetDelegateRun(func(ctx context.Context, cmd *exec.Command) error {
			require.Equal(t, dockerCmd, cmd.Name)
			require.Equal(t, []string{"--config", "test_config_dir", "run", "--volume", "/tmp/test:/OUT", "--env", "SKIP_BUILD=1", "https://gcr.io/skia-public/skia-release:123", "test_cmd"}, cmd.Args)
			return nil
		})
		ctx = td.WithExecRunFn(ctx, mockRun.Run)

		err := Run(ctx, "https://gcr.io/skia-public/skia-release:123", "test_config_dir", []string{"test_cmd"}, []string{"/tmp/test:/OUT"}, []string{"SKIP_BUILD=1"})
		require.NoError(t, err)

		return nil
	})
}

func TestPull(t *testing.T) {
	unittest.SmallTest(t)

	_ = td.RunTestSteps(t, false, func(ctx context.Context) error {
		mockRun := &exec.CommandCollector{}
		mockRun.SetDelegateRun(func(ctx context.Context, cmd *exec.Command) error {
			require.Equal(t, dockerCmd, cmd.Name)
			require.Equal(t, []string{"--config", "test_config_dir", "pull", "https://gcr.io/skia-public/skia-release:123"}, cmd.Args)
			require.Equal(t, "", cmd.Dir)
			return nil
		})
		ctx = td.WithExecRunFn(ctx, mockRun.Run)

		err := Pull(ctx, "https://gcr.io/skia-public/skia-release:123", "test_config_dir")
		require.NoError(t, err)

		return nil
	})
}

func TestPush(t *testing.T) {
	unittest.SmallTest(t)

	_ = td.RunTestSteps(t, false, func(ctx context.Context) error {
		mockRun := &exec.CommandCollector{}
		mockRun.SetDelegateRun(func(ctx context.Context, cmd *exec.Command) error {
			require.Equal(t, dockerCmd, cmd.Name)
			require.Equal(t, []string{"--config", "test_config_dir", "push", "https://gcr.io/skia-public/skia-release:123"}, cmd.Args)
			require.Equal(t, ".", cmd.Dir)
			_, err := cmd.CombinedOutput.Write([]byte(`The push refers to repository [gcr.io/skia-public/linux-run]
d75098a9b75c: Preparing
22f92a22a0a1: Preparing
7c18f43554d6: Preparing
c90191647a49: Preparing
22f92a22a0a1: Layer already exists
c90191647a49: Layer already exists
7c18f43554d6: Layer already exists
d75098a9b75c: Layer already exists
latest: digest: sha256:9f856122da361de5a737c4dd5b0ef582df3e051e6586d971a068e72657ddd0d2 size: 1166`))
			return err
		})
		ctx = td.WithExecRunFn(ctx, mockRun.Run)

		sha256, err := Push(ctx, "https://gcr.io/skia-public/skia-release:123", "test_config_dir")
		require.NoError(t, err)
		require.Equal(t, "sha256:9f856122da361de5a737c4dd5b0ef582df3e051e6586d971a068e72657ddd0d2", sha256)

		return nil
	})
}
