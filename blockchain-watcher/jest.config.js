/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  moduleFileExtensions: ["js", "json", "ts"],
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
      lines: 55,
    },
  },
};
