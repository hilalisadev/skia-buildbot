import * as paramset from './index.js'

describe('ParamSet',
  function() {
    function testParamSet() {
      let ps = {};
      // The empty paramset matches everything.
      assert.isTrue(paramset.match(ps, {}));
      assert.isTrue(paramset.match(ps, {"foo": "2", "bar": "a"}));

      let p = {
        "foo": "1",
        "bar": "a",
      };
      paramset.add(ps, p);
      assert.isTrue(paramset.match(ps, p));
      assert.isFalse(paramset.match(ps, {}));

      paramset.add(ps, {
        "foo": "2",
        "bar": "b",
      });

      paramset.add(ps, {
        "foo": "1",
        "bar": "b",
      });

      assert.isTrue(paramset.match(ps,  {"foo": "2", "bar": "a"}));
      assert.isTrue(paramset.match(ps,  {"foo": "2", "bar": "a", "baz": "other"}));
      assert.isFalse(paramset.match(ps, {            "bar": "a"}));
      assert.isFalse(paramset.match(ps, {"foo": "2"            }));
      assert.isFalse(paramset.match(ps, {                      }));
      assert.isFalse(paramset.match(ps, {"foo": "3", "bar": "a"}));
      assert.isFalse(paramset.match(ps, {"foo": "2", "bar": "c"}));
    }

    function testParamSetWithIgnore() {
      let ps = {}
      let p = {
        "foo": "1",
        "bar": "a",
        "description": "long rambling text",
      };
      paramset.add(ps, p, ['description']);
      assert.isTrue(paramset.match(ps, p));
      assert.isTrue(paramset.match(ps, {
        "foo": "1",
        "bar": "a",
      }));
      assert.isFalse(paramset.match(ps, {}));
    }


    it('should be able get match against params', function() {
      testParamSet();
      testParamSetWithIgnore();
    });
  }
);