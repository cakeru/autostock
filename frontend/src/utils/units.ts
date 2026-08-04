export type DistanceUnit = 'km' | 'mi'

// The shop picks one distance system (km or mi) in Settings; all odometer and
// interval values are entered and displayed in that unit.
export function distanceUnit(settings?: { distance_unit?: string }): DistanceUnit {
  return settings?.distance_unit === 'mi' ? 'mi' : 'km'
}

export function unitLabel(unit: DistanceUnit): string {
  return unit === 'mi' ? 'mi' : 'km'
}

export function unitWord(unit: DistanceUnit): string {
  return unit === 'mi' ? 'Miles' : 'Kilometers'
}

// Storage is always km (backend convention); these convert to/from the shop's
// display unit so a miles-based shop enters and reads values in miles.
const KM_PER_MI = 1.609344

export function kmToMi(km: number): number {
  return Math.round(km / KM_PER_MI)
}

export function miToKm(mi: number): number {
  return Math.round(mi * KM_PER_MI)
}
