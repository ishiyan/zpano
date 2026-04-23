import { Adaptivity } from './adaptivity.js';
import { IndicatorIdentifier } from './indicator-identifier.js';
import { InputRequirement } from './input-requirement.js';
import { OutputDescriptor } from './output-descriptor.js';
import { VolumeUsage } from './volume-usage.js';

/**
 * Classifies an indicator along multiple taxonomic dimensions to enable
 * filtering and display in charting catalogs.
 */
export interface Descriptor {
  /** Uniquely identifies the indicator. */
  identifier: IndicatorIdentifier;

  /** Groups related indicators (e.g., by author or category). */
  family: string;

  /** Whether the indicator adapts its parameters. */
  adaptivity: Adaptivity;

  /** The minimum input data type this indicator consumes. */
  inputRequirement: InputRequirement;

  /** How this indicator uses volume information. */
  volumeUsage: VolumeUsage;

  /** Per-output classifications. */
  outputs: OutputDescriptor[];
}
