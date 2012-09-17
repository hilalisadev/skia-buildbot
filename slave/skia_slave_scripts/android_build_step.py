# Copyright (c) 2012 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

""" Subclass for all slave-side Android build steps. """

from build_step import BuildStep
from utils import misc

class AndroidBuildStep(BuildStep):
  def _PreRun(self):
    misc.RunADB(self._serial, ['root'])
    misc.RunADB(self._serial, ['remount'])
    misc.SetCPUScalingMode(self._serial, 'performance')

  def __init__(self, args, attempts=1):
    self._device = args['device']
    self._serial = args['serial']
    if self._serial == 'None':
      self._serial = misc.GetSerial(self._device)
    self._android_scratch_dir = '%s/skiabot' % misc.BashGet(
        "%s -s %s shell echo \$EXTERNAL_STORAGE" %
            (misc.PATH_TO_ADB, self._serial),
        echo=True).rstrip()
    self._android_gm_dir = '%s/skia_gm' % self._android_scratch_dir
    self._android_perf_dir = '%s/skia_perf' % self._android_scratch_dir
    self._android_skp_dir = '%s/skia_skp' % self._android_scratch_dir
    self._android_skp_perf_dir = '%s/skia_skp_perf' % self._android_scratch_dir
    self._android_skp_out_dir = '%s/skia_skp_out' % self._android_scratch_dir
    super(AndroidBuildStep, self).__init__(args, attempts)