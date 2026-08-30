import { defineConfig } from "vitest/config";
export default defineConfig({
  test: {
    // Isolate environment per test file
    isolate: true,
    // Fail fast if a test fails in CI
    bail: process.env.CI ? 1 : 0,
    // Coverage reporting
    coverage: {
      provider: "v8", // or 'istanbul'
      reporter: ["text", "json", "html", "lcov"],
      reportsDirectory: "./coverage",
      exclude: ["node_modules/", "dist/", "**/*.d.ts", "**/vitest.config.*"],
    },
    // Global setup files (e.g. DB mocks, env vars)
    setupFiles: ["./src/test/setup.ts"],
    // Clear mocks between tests
    clearMocks: true,
    restoreMocks: true,
    // Timeout per test
    testTimeout: 5000,
    // Snapshot directory
    snapshotFormat: {
      indent: 2,
      printBasicPrototype: true,
    },
    include: ["tests/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}"],
    environment: "jsdom",
    // Reporters
    reporters: process.env.CI ? ["default", "junit"] : ["default"],
    outputFile: process.env.CI ? "./test-results/junit.xml" : undefined,
  },
});
