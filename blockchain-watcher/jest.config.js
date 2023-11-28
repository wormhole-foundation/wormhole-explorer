/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  moduleFileExtensions: ["js", "json", "ts"],
  setupFiles: ["<rootDir>/src/infrastructure/log.ts"],
  roots: ["test", "src"],
  testRegex: ".*\\.test\\.ts$",
  transform: {
    "^.+\\.(t|j)s$": "ts-jest",
  },
  collectCoverage: true,
  collectCoverageFrom: ["**/*.(t|j)s"],
  coveragePathIgnorePatterns: ["node_modules", "test"],
  coverageDirectory: "./coverage",
  coverageThreshold: {
    global: {
      lines: 58.5,
    },
  },
};
