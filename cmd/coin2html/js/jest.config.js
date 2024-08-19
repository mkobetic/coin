/** @type {import('ts-jest/dist/types').InitialOptionsTsJest} */
module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  moduleDirectories: ["node_modules"],
  transformIgnorePatterns: [`node_modules`],
  moduleNameMapper: {
    d3: "<rootDir>/node_modules/d3/dist/d3.min.js",
  },
};
