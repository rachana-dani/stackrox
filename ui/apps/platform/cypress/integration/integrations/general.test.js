import { selectors } from '../../constants/IntegrationsPage';
import withAuth from '../../helpers/basicAuth';

const toastShouldContain = (message) => {
    cy.get('.toast-selector').should('contain', message).find('button').click({ force: true });
};

describe('Integrations page', () => {
    withAuth();

    beforeEach(() => {
        cy.visit('/');
        cy.get(selectors.configure).click();
        cy.get(selectors.navLink).click({ force: true });
    });

    it('Plugin tiles should all be the same height', () => {
        let value = null;
        cy.get(selectors.plugins).each(($el) => {
            if (value) {
                expect($el[0].clientHeight).to.equal(value);
            } else {
                value = $el[0].clientHeight;
            }
        });
    });

    it.skip('should test, create, update, and delete an integration with Slack', () => {
        cy.get(selectors.slackTile).click();
        cy.get(selectors.buttons.delete).should('not.exist');
        cy.get(selectors.buttons.new).click();

        // test that validation error happens when form is incomplete
        cy.get(selectors.buttons.test).click();
        toastShouldContain('error');

        const nameInput = `Slack Test ${Math.random().toString(36).substring(7)}`;
        const defaultWebhook = 'https://hooks.slack.com/services/EXAMPLE';
        const labelAnnotationKey = 'slack-test';

        cy.get(selectors.slackForm.nameInput).type(nameInput);
        cy.get(selectors.slackForm.defaultWebhook).type(defaultWebhook);
        cy.get(selectors.slackForm.labelAnnotationKey).type(labelAnnotationKey);

        // the test button should not return an error with valid inputs
        cy.get(selectors.buttons.test).click();
        toastShouldContain('Integration test was successful');

        // test creating an integration
        cy.get(selectors.buttons.create).click({ force: true });
        toastShouldContain('Successfully integrated slack');

        // test updating an existing integration
        cy.get(`${selectors.table.rows}:contains("${nameInput}")`).click();
        cy.get(selectors.buttons.save).click({ force: true });
        cy.get(selectors.buttons.closePanel).click({ force: true });
        toastShouldContain('Successfully integrated slack');

        // test deleting an integration
        cy.get(`${selectors.table.rows}:contains("${nameInput}") input[type="checkbox"]`).check();
        cy.get(selectors.buttons.delete).click({ force: true });
        cy.get(selectors.buttons.confirm).click();
        toastShouldContain('Successfully deleted 1 integration');
        cy.get(`${selectors.table.rows}:contains("${nameInput}")`).should('not.exist');
    });

    it.skip('should test, create, update, and delete an integration with DockerHub', () => {
        cy.get(selectors.dockerRegistryTile).click();
        cy.get(selectors.buttons.delete).should('not.exist');
        cy.get(selectors.buttons.new).click();

        const nameInput = `Docker Registry ${Math.random().toString(36).substring(7)}`;
        const endpointInput = 'registry-1.docker.io';

        cy.get(selectors.dockerRegistryForm.nameInput).type(nameInput);

        cy.get(
            `${selectors.dockerRegistryForm.typesSelect} .react-select__dropdown-indicator`
        ).click();
        cy.get('.react-select__menu-list > div:contains("Registry")').click();

        // test that validation error happens when form is incomplete
        cy.get(selectors.buttons.test).click();
        toastShouldContain('error');

        cy.get(selectors.dockerRegistryForm.endpointInput).type(endpointInput);

        // the test button should not return an error with valid inputs
        cy.get(selectors.buttons.test).click();
        toastShouldContain('Integration test was successful');

        // test creating an integration
        cy.get(selectors.buttons.create).click({ force: true });
        toastShouldContain('Successfully integrated docker');

        // test updating an existing integration
        cy.get(`${selectors.table.rows}:contains("${nameInput}")`).click();
        cy.get(selectors.buttons.save).click({ force: true });
        cy.get(selectors.buttons.closePanel).click({ force: true });
        toastShouldContain('Successfully integrated docker');

        // test deleting an integration
        cy.get(`${selectors.table.rows}:contains("${nameInput}") input[type="checkbox"]`).check();
        cy.get(selectors.buttons.delete).click({ force: true });
        cy.get(selectors.buttons.confirm).click();
        toastShouldContain('Successfully deleted 1 integration');
        cy.get(`${selectors.table.rows}:contains("${nameInput}")`).should('not.exist');
    });
});

describe('API Token Creation Flow', () => {
    withAuth();

    const randomTokenName = `Token${Math.random().toString(36).substring(7)}`;

    beforeEach(() => {
        cy.visit('/');
        cy.get(selectors.configure).click();
        cy.get(selectors.navLink).click({ force: true });
    });

    it('should show table for API Tokens', () => {
        cy.get(selectors.apiTokenTile).click();
        cy.get('.pf-c-breadcrumb').contains('apitoken');
    });

    it('should be able to generate an API token', () => {
        cy.get(selectors.apiTokenTile).click();
        cy.get(selectors.buttons.newIntegration).click();
        cy.get(selectors.apiTokenForm.nameInput).type(randomTokenName);
        cy.get(`${selectors.apiTokenForm.roleSelect} .react-select__dropdown-indicator`).click();
        cy.get('.react-select__menu-list > div:contains("Admin")').click();
        cy.get(selectors.buttons.generate).click();
        cy.get(selectors.apiTokenBox);
        cy.get(selectors.apiTokenDetailsDiv).contains(`Name:${randomTokenName}`);
        cy.get(selectors.apiTokenDetailsDiv).contains('Roles:Admin');
    });

    it('should show the generated API token in the table, and be clickable', () => {
        cy.get(selectors.apiTokenTile).click();
        cy.get(`td:contains("${randomTokenName}") button`).click();
        cy.get(selectors.apiTokenDetailsDiv).contains(`Name:${randomTokenName}`);
        cy.get(selectors.apiTokenDetailsDiv).contains('Roles:Admin');
    });

    it('should be able to revoke the API token', () => {
        cy.get(selectors.apiTokenTile).click();
        cy.get(`tr:contains("${randomTokenName}") input`).check();
        cy.get(selectors.buttons.delete).click({ force: true });
        cy.get(selectors.buttons.confirm).click();
        cy.get(`td:contains("${randomTokenName}")`).should('not.exist');
    });
});
