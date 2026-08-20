// Package thermo carries the physical constants and the Arrhenius rate
// parameters shared by the reactor models. Every rate-law evaluation inside
// batch-arrh uses the same gas constant and the same temperature convention
// (absolute kelvin), so a k computed here is interchangeable with a k given
// directly in the scenario file.
package thermo

// R is the universal gas constant in J/(mol*K). All activation energies in
// the scenario files are expressed in J/mol and all temperatures in kelvin,
// which keeps the exponent dimensionless: Ea/(R*T).
const R = 8.31446261815324

// KelvinAtZeroC is the offset between the celsius and kelvin scales.
const KelvinAtZeroC = 273.15

// JoulesPerKilojoule is the factor that converts kJ/mol to J/mol. Scenario
// files may state activation energies in either unit; the converters below
// keep the two spellings consistent inside one model.
const JoulesPerKilojoule = 1000.0
