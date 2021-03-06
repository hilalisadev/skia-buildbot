import './index.js';
import { setUpElementUnderTest } from '../test_util';
import { $$ } from 'common-sk/modules/dom';

describe('edit-ignore-rule-sk', () => {
  const newInstance = setUpElementUnderTest('edit-ignore-rule-sk');

  // This date is arbitrary
  const fakeNow = Date.parse('2020-02-01T00:00:00Z');
  const regularNow = Date.now;

  let editIgnoreRuleSk;
  beforeEach(() => {
    editIgnoreRuleSk = newInstance();
    // All tests will have the paramset loaded.
    editIgnoreRuleSk.paramset = {
      'alpha_type': ['Opaque', 'Premul'],
      'arch': ['arm', 'arm64', 'x86', 'x86_64'],
    };
    Date.now = () => fakeNow;
  });

  afterEach(() => {
    Date.now = regularNow;
  });

  describe('inputs and outputs', () => {
    it('has no query, note or expires', () => {
      expect(editIgnoreRuleSk.query).to.equal('');
      expect(editIgnoreRuleSk.note).to.equal('');
      expect(editIgnoreRuleSk.expires).to.equal('');
    });

    it('reflects typed in values', () => {
      getExpiresInput(editIgnoreRuleSk).value = '2w';
      getNoteInput(editIgnoreRuleSk).value = 'this is a bug';
      expect(editIgnoreRuleSk.expires).to.equal('2w');
      expect(editIgnoreRuleSk.note).to.equal('this is a bug');
    });

    it('reflects interactions with the query-sk element', () => {
      // Select alpha_type key, which displays Opaque and Premul as values.
      getFirstQuerySkKey(editIgnoreRuleSk).click();
      // Select Opaque as a value
      getFirstQuerySkValue(editIgnoreRuleSk).click();
      expect(editIgnoreRuleSk.query).to.equal('alpha_type=Opaque');
    });

    it('converts future dates to human readable durations', () => {
      editIgnoreRuleSk.expires = '2020-02-07T06:00:00Z';
      // It is ok that the 6 hours gets rounded out.
      expect(editIgnoreRuleSk.expires).to.equal('6d');
    });

    it('converts past or invalid dates to nothing (requiring them to be re-input)', () => {
      editIgnoreRuleSk.expires = '2020-01-07T06:00:00Z';
      expect(editIgnoreRuleSk.expires).to.equal('');
      editIgnoreRuleSk.expires = 'invalid date';
      expect(editIgnoreRuleSk.expires).to.equal('');
    });
  });

  describe('validation', () => {
    it('has the error msg hidden by default', () => {
      expect(getErrorMessage(editIgnoreRuleSk).hasAttribute('hidden')).to.be.true;
    });

    it('does not validate when query is empty', () => {
      editIgnoreRuleSk.query = '';
      editIgnoreRuleSk.expires = '2w';
      expect(editIgnoreRuleSk.verifyFields()).to.be.false;
      expect(getErrorMessage(editIgnoreRuleSk).hasAttribute('hidden')).to.be.false;
    });

    it('does not validate when expires is empty', () => {
      editIgnoreRuleSk.query = 'alpha_type=Opaque';
      editIgnoreRuleSk.expires = '';
      expect(editIgnoreRuleSk.verifyFields()).to.be.false;
      expect(getErrorMessage(editIgnoreRuleSk).hasAttribute('hidden')).to.be.false;
    });

    it('does not validate when both expires and query are empty', () => {
      editIgnoreRuleSk.query = '';
      editIgnoreRuleSk.expires = '';
      expect(editIgnoreRuleSk.verifyFields()).to.be.false;
      expect(getErrorMessage(editIgnoreRuleSk).hasAttribute('hidden')).to.be.false;
    });

    it('does passes validation when both expires and query are set', () => {
      editIgnoreRuleSk.query = 'foo=bar';
      getExpiresInput(editIgnoreRuleSk).value = '1w';
      expect(editIgnoreRuleSk.verifyFields()).to.be.true;
      expect(getErrorMessage(editIgnoreRuleSk).hasAttribute('hidden')).to.be.true;
    });
  });
});

const getExpiresInput = (ele) => $$('#expires', ele);

const getNoteInput = (ele) => $$('#note', ele);

const getErrorMessage = (ele) => $$('.error', ele);

const getFirstQuerySkKey = (ele) => $$('query-sk .selection div:nth-child(1)', ele);

const getFirstQuerySkValue = (ele) => $$('query-sk #values div:nth-child(1)', ele);
