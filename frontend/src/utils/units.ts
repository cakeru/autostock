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
