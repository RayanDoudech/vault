/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: BUSL-1.1
 */

import { format } from 'd3-format';
import { mean } from 'd3-array';

// COLOR THEME:
export const BLUE_PALETTE = ['#cce3fe', '#0c56e9', '#1c345f']; // blues from https://helios.hashicorp.design/foundations/colors?tab=palette#core-palette
export const UPGRADE_WARNING = '#FDEEBA';
export const GREY = '#EBEEF2';

// TRANSLATIONS:
export const TRANSLATE = { left: -11 };
export const SVG_DIMENSIONS = { height: 190, width: 500 };

// Reference for tickFormat https://www.youtube.com/watch?v=c3MCROTNN8g
export function formatNumbers(number) {
  if (number < 1000) return number;
  if (number < 10000) return format('.1s')(number);
  // replace SI prefix of 'G' for billions to 'B'
  return format('.2s')(number).replace('G', 'B');
}

export function calculateAverage(dataset, objectKey) {
  // before mapping for values, check that the objectKey exists at least once in the dataset because
  // map returns 0 when dataset[objectKey] is undefined in order to calculate average
  if (!Array.isArray(dataset) || !objectKey || !dataset.some((d) => Object.keys(d).includes(objectKey))) {
    return null;
  }

  const integers = dataset.map((d) => (d[objectKey] ? d[objectKey] : 0));
  const checkIntegers = integers.every((n) => Number.isInteger(n)); // decimals will be false
  return checkIntegers ? Math.round(mean(integers)) : null;
}

export function calculateSum(integerArray) {
  if (!Array.isArray(integerArray) || integerArray.some((n) => typeof n !== 'number')) {
    return null;
  }
  return integerArray.reduce((a, b) => a + b, 0);
}
