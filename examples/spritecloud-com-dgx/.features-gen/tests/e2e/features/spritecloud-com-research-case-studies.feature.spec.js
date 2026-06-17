// Generated from: tests/e2e/features/spritecloud-com-research-case-studies.feature
import { test } from "playwright-bdd";

test.describe('WwwSpritecloudCom — research journey', () => {

  test('research journey reaches its terminal page', { tag: ['@journey:research', '@priority:standard', '@smoke'] }, async ({ Given, When, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await And('the page title contains "spriteCloud - Test your software, not your reputation!"', null, { page }); 
    await And('the main heading reads "Test your software, not your reputation."', null, { page }); 
    await When('I click the link to "/case-studies"', null, { page }); 
    await And('the page title contains "Case Studies"', null, { page }); 
  });

  test('research — deep-link to the terminal page renders correctly', { tag: ['@journey:research', '@priority:standard', '@kind:resume'] }, async ({ Given, page }) => { 
    await Given('I open the page "/case-studies"', null, { page }); 
  });

  test('research — back button after navigation returns to landing', { tag: ['@journey:research', '@priority:standard', '@kind:back-button'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/case-studies"', null, { page }); 
    await When('I go back in the browser history', null, { page }); 
    await Then('the main heading reads "Test your software, not your reputation."', null, { page }); 
  });

});

// == technical section ==

test.use({
  $test: [({}, use) => use(test), { scope: 'test', box: true }],
  $uri: [({}, use) => use('tests/e2e/features/spritecloud-com-research-case-studies.feature'), { scope: 'test', box: true }],
  $bddFileData: [({}, use) => use(bddFileData), { scope: "test", box: true }],
});

const bddFileData = [ // bdd-data-start
  {"pwTestLine":6,"pickleLine":20,"tags":["@journey:research","@priority:standard","@smoke"],"steps":[{"pwStepLine":7,"gherkinStepLine":21,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":8,"gherkinStepLine":22,"keywordType":"Context","textWithKeyword":"And the page title contains \"spriteCloud - Test your software, not your reputation!\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Test your software, not your reputation!"}}]},{"pwStepLine":9,"gherkinStepLine":23,"keywordType":"Context","textWithKeyword":"And the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]},{"pwStepLine":10,"gherkinStepLine":24,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-studies\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-studies"}}]},{"pwStepLine":11,"gherkinStepLine":25,"keywordType":"Action","textWithKeyword":"And the page title contains \"Case Studies\"","stepMatchArguments":[{"group":{"start":25,"value":"Case Studies"}}]}]},
  {"pwTestLine":14,"pickleLine":28,"tags":["@journey:research","@priority:standard","@kind:resume"],"steps":[{"pwStepLine":15,"gherkinStepLine":29,"keywordType":"Context","textWithKeyword":"Given I open the page \"/case-studies\"","stepMatchArguments":[{"group":{"start":17,"value":"/case-studies"}}]}]},
  {"pwTestLine":18,"pickleLine":32,"tags":["@journey:research","@priority:standard","@kind:back-button"],"steps":[{"pwStepLine":19,"gherkinStepLine":33,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":20,"gherkinStepLine":34,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-studies\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-studies"}}]},{"pwStepLine":21,"gherkinStepLine":35,"keywordType":"Action","textWithKeyword":"When I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":22,"gherkinStepLine":36,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]}]},
]; // bdd-data-end