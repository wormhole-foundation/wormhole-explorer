/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  collectCoverageFrom: [
    "./src/domain",
    "./src/infrastructure/mappers",
    "./src/infrastructure/repositories",
  ],
  coverageThreshold: {
    global: {
      lines: 85,
    },
  },
};
