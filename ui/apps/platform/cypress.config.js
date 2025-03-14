/*
 * The helper function intended to provide automatic code completion for configuration in many popular code editors
 * had subtle side-effect to cause some typescript-eslint/no-unsafe-return errors in unit test files.
 *
 * const { defineConfig } = require('cypress'); // eslint-disable-line import/no-extraneous-dependencies
 * module.exports = defineConfig({ … });
 */

module.exports = {
    blockHosts: ['*.*'], // Browser options
    chromeWebSecurity: false, // Browser options
    numTestsKeptInMemory: 0, // Global options
    requestTimeout: 10000, // Timeouts options
    viewportHeight: 850, // Viewport options
    viewportWidth: 1440, // Viewport options

    e2e: {
        baseUrl: 'https://localhost:3000',
        specPattern: 'cypress/integration/**/*.test.js',
    },
};
