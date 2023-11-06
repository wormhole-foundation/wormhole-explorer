/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  testRegex: "^(?!.*integration.*)(?=.*test\\/).*\\.test\\.ts$",
  collectCoverageFrom: ["./src/**"],
  coverageThreshold: {
    global: {
      lines: 85,
    },
  },
};
