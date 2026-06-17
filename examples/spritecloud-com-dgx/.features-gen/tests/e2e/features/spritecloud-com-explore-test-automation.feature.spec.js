// Generated from: tests/e2e/features/spritecloud-com-explore-test-automation.feature
import { test } from "playwright-bdd";

test.describe('WwwSpritecloudCom — explore journey', () => {

  test('explore journey reaches its terminal page', { tag: ['@journey:explore', '@priority:nice-to-have', '@smoke'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await And('the page title contains "spriteCloud - Test your software, not your reputation!"', null, { page }); 
    await And('the main heading reads "Test your software, not your reputation."', null, { page }); 
    await When('I click the link to "/test-automation"', null, { page }); 
    await Then('I see the heading "Expert Test Automation Services"', null, { page }); 
    await And('the page title contains "Test Automation"', null, { page }); 
  });

  test('explore — deep-link to the terminal page renders correctly', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:resume'] }, async ({ Given, Then, page }) => { 
    await Given('I open the page "/test-automation"', null, { page }); 
    await Then('I see the heading "Expert Test Automation Services"', null, { page }); 
  });

  test('explore — back button after navigation returns to landing', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:back-button'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/test-automation"', null, { page }); 
    await When('I go back in the browser history', null, { page }); 
    await Then('the main heading reads "Test your software, not your reputation."', null, { page }); 
  });

  test('explore — switching to landing and back leaves no broken state', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:cross-journey'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I navigate directly to "/"', null, { page }); 
    await And('I go back in the browser history', null, { page }); 
    await Then('no error message is shown in the form region', null, { page }); 
  });

  test('Submit form twice in rapid succession', { tag: ['@journey:explore', '@priority:nice-to-have', '@llm-composed', '@kind:edge', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I am on the landing page', null, { page }); 
    await When('I submit the form twice in rapid succession', null, { page }); 
    await Then('the form is not double-submitted', null, { page }); 
  });

  test('Fill, submit, then reload mid-flow', { tag: ['@journey:explore', '@priority:nice-to-have', '@llm-composed', '@kind:edge', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I am on the landing page', null, { page }); 
    await When('I enter "test@example.com" into the "email" field', null, { page }); 
    await Then('the "email" field has the value "test@example.com"', null, { page }); 
    await When('I reload the page', null, { page }); 
    await Then('the "email" field has the value "test@example.com"', null, { page }); 
  });

});

// == technical section ==

test.use({
  $test: [({}, use) => use(test), { scope: 'test', box: true }],
  $uri: [({}, use) => use('tests/e2e/features/spritecloud-com-explore-test-automation.feature'), { scope: 'test', box: true }],
  $bddFileData: [({}, use) => use(bddFileData), { scope: "test", box: true }],
});

const bddFileData = [ // bdd-data-start
  {"pwTestLine":6,"pickleLine":20,"tags":["@journey:explore","@priority:nice-to-have","@smoke"],"steps":[{"pwStepLine":7,"gherkinStepLine":21,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":8,"gherkinStepLine":22,"keywordType":"Context","textWithKeyword":"And the page title contains \"spriteCloud - Test your software, not your reputation!\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Test your software, not your reputation!"}}]},{"pwStepLine":9,"gherkinStepLine":23,"keywordType":"Context","textWithKeyword":"And the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]},{"pwStepLine":10,"gherkinStepLine":24,"keywordType":"Action","textWithKeyword":"When I click the link to \"/test-automation\"","stepMatchArguments":[{"group":{"start":21,"value":"/test-automation"}}]},{"pwStepLine":11,"gherkinStepLine":25,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Expert Test Automation Services\"","stepMatchArguments":[{"group":{"start":19,"value":"Expert Test Automation Services"}}]},{"pwStepLine":12,"gherkinStepLine":26,"keywordType":"Outcome","textWithKeyword":"And the page title contains \"Test Automation\"","stepMatchArguments":[{"group":{"start":25,"value":"Test Automation"}}]}]},
  {"pwTestLine":15,"pickleLine":29,"tags":["@journey:explore","@priority:nice-to-have","@kind:resume"],"steps":[{"pwStepLine":16,"gherkinStepLine":30,"keywordType":"Context","textWithKeyword":"Given I open the page \"/test-automation\"","stepMatchArguments":[{"group":{"start":17,"value":"/test-automation"}}]},{"pwStepLine":17,"gherkinStepLine":31,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Expert Test Automation Services\"","stepMatchArguments":[{"group":{"start":19,"value":"Expert Test Automation Services"}}]}]},
  {"pwTestLine":20,"pickleLine":34,"tags":["@journey:explore","@priority:nice-to-have","@kind:back-button"],"steps":[{"pwStepLine":21,"gherkinStepLine":35,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":22,"gherkinStepLine":36,"keywordType":"Action","textWithKeyword":"When I click the link to \"/test-automation\"","stepMatchArguments":[{"group":{"start":21,"value":"/test-automation"}}]},{"pwStepLine":23,"gherkinStepLine":37,"keywordType":"Action","textWithKeyword":"When I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":24,"gherkinStepLine":38,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]}]},
  {"pwTestLine":27,"pickleLine":41,"tags":["@journey:explore","@priority:nice-to-have","@kind:cross-journey"],"steps":[{"pwStepLine":28,"gherkinStepLine":42,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":29,"gherkinStepLine":43,"keywordType":"Action","textWithKeyword":"When I navigate directly to \"/\"","stepMatchArguments":[{"group":{"start":24,"value":"/"}}]},{"pwStepLine":30,"gherkinStepLine":44,"keywordType":"Action","textWithKeyword":"And I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":31,"gherkinStepLine":45,"keywordType":"Outcome","textWithKeyword":"Then no error message is shown in the form region","stepMatchArguments":[]}]},
  {"pwTestLine":34,"pickleLine":53,"tags":["@journey:explore","@priority:nice-to-have","@llm-composed","@kind:edge","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":35,"gherkinStepLine":54,"keywordType":"Context","textWithKeyword":"Given I am on the landing page","stepMatchArguments":[]},{"pwStepLine":36,"gherkinStepLine":55,"keywordType":"Action","textWithKeyword":"When I submit the form twice in rapid succession","stepMatchArguments":[]},{"pwStepLine":37,"gherkinStepLine":56,"keywordType":"Outcome","textWithKeyword":"Then the form is not double-submitted","stepMatchArguments":[]}]},
  {"pwTestLine":40,"pickleLine":59,"tags":["@journey:explore","@priority:nice-to-have","@llm-composed","@kind:edge","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":41,"gherkinStepLine":60,"keywordType":"Context","textWithKeyword":"Given I am on the landing page","stepMatchArguments":[]},{"pwStepLine":42,"gherkinStepLine":61,"keywordType":"Action","textWithKeyword":"When I enter \"test@example.com\" into the \"email\" field","stepMatchArguments":[{"group":{"start":9,"value":"test@example.com"}},{"group":{"start":37,"value":"email"}}]},{"pwStepLine":43,"gherkinStepLine":62,"keywordType":"Outcome","textWithKeyword":"Then the \"email\" field has the value \"test@example.com\"","stepMatchArguments":[{"group":{"start":5,"value":"email"}},{"group":{"start":33,"value":"test@example.com"}}]},{"pwStepLine":44,"gherkinStepLine":63,"keywordType":"Action","textWithKeyword":"When I reload the page","stepMatchArguments":[]},{"pwStepLine":45,"gherkinStepLine":64,"keywordType":"Outcome","textWithKeyword":"Then the \"email\" field has the value \"test@example.com\"","stepMatchArguments":[{"group":{"start":5,"value":"email"}},{"group":{"start":33,"value":"test@example.com"}}]}]},
]; // bdd-data-end