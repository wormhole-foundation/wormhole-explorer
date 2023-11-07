/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  collectCoverageFrom: ["./src/**"],
  coverageThreshold: {
    global: {
      lines: 85,
    },
  },
};
