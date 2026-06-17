// Generated from: tests/e2e/features/spritecloud-com-explore-cybersecurity.feature
import { test } from "playwright-bdd";

test.describe('WwwSpritecloudCom — explore journey', () => {

  test('explore journey reaches its terminal page', { tag: ['@journey:explore', '@priority:nice-to-have', '@smoke'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await And('the page title contains "spriteCloud - Test your software, not your reputation!"', null, { page }); 
    await And('the main heading reads "Test your software, not your reputation."', null, { page }); 
    await When('I click the link to "/cybersecurity"', null, { page }); 
    await Then('I see the heading "Fully Tailored Cybersecurity."', null, { page }); 
    await And('the page title contains "spriteCloud - Penetration Testing"', null, { page }); 
  });

  test('explore — deep-link to the terminal page renders correctly', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:resume'] }, async ({ Given, Then, page }) => { 
    await Given('I open the page "/cybersecurity"', null, { page }); 
    await Then('I see the heading "Fully Tailored Cybersecurity."', null, { page }); 
  });

  test('explore — back button after navigation returns to landing', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:back-button'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/cybersecurity"', null, { page }); 
    await When('I go back in the browser history', null, { page }); 
    await Then('the main heading reads "Test your software, not your reputation."', null, { page }); 
  });

  test('explore — switching to landing and back leaves no broken state', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:cross-journey'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I navigate directly to "/"', null, { page }); 
    await And('I go back in the browser history', null, { page }); 
    await Then('no error message is shown in the form region', null, { page }); 
  });

  test('Navigate to cybersecurity page and verify title and heading', { tag: ['@journey:explore', '@priority:nice-to-have', '@llm-composed', '@kind:happy', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/cybersecurity"', null, { page }); 
    await Then('the page title contains "Penetration Testing"', null, { page }); 
    await Then('the main heading reads "Fully Tailored Cybersecurity."', null, { page }); 
  });

  test('Direct navigation to cybersecurity page', { tag: ['@journey:explore', '@priority:nice-to-have', '@llm-composed', '@kind:edge', '@model:qwen3-coder-next-latest'] }, async ({ Given, Then, page }) => { 
    await Given('I open the page "/cybersecurity"', null, { page }); 
    await Then('the page title contains "Penetration Testing"', null, { page }); 
    await Then('the main heading reads "Fully Tailored Cybersecurity."', null, { page }); 
  });

  test('Reload cybersecurity page preserves heading', { tag: ['@journey:explore', '@priority:nice-to-have', '@llm-composed', '@kind:edge', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the page "/cybersecurity"', null, { page }); 
    await When('I reload the page', null, { page }); 
    await Then('the main heading reads "Fully Tailored Cybersecurity."', null, { page }); 
  });

});

// == technical section ==

test.use({
  $test: [({}, use) => use(test), { scope: 'test', box: true }],
  $uri: [({}, use) => use('tests/e2e/features/spritecloud-com-explore-cybersecurity.feature'), { scope: 'test', box: true }],
  $bddFileData: [({}, use) => use(bddFileData), { scope: "test", box: true }],
});

const bddFileData = [ // bdd-data-start
  {"pwTestLine":6,"pickleLine":20,"tags":["@journey:explore","@priority:nice-to-have","@smoke"],"steps":[{"pwStepLine":7,"gherkinStepLine":21,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":8,"gherkinStepLine":22,"keywordType":"Context","textWithKeyword":"And the page title contains \"spriteCloud - Test your software, not your reputation!\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Test your software, not your reputation!"}}]},{"pwStepLine":9,"gherkinStepLine":23,"keywordType":"Context","textWithKeyword":"And the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]},{"pwStepLine":10,"gherkinStepLine":24,"keywordType":"Action","textWithKeyword":"When I click the link to \"/cybersecurity\"","stepMatchArguments":[{"group":{"start":21,"value":"/cybersecurity"}}]},{"pwStepLine":11,"gherkinStepLine":25,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Fully Tailored Cybersecurity.\"","stepMatchArguments":[{"group":{"start":19,"value":"Fully Tailored Cybersecurity."}}]},{"pwStepLine":12,"gherkinStepLine":26,"keywordType":"Outcome","textWithKeyword":"And the page title contains \"spriteCloud - Penetration Testing\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Penetration Testing"}}]}]},
  {"pwTestLine":15,"pickleLine":29,"tags":["@journey:explore","@priority:nice-to-have","@kind:resume"],"steps":[{"pwStepLine":16,"gherkinStepLine":30,"keywordType":"Context","textWithKeyword":"Given I open the page \"/cybersecurity\"","stepMatchArguments":[{"group":{"start":17,"value":"/cybersecurity"}}]},{"pwStepLine":17,"gherkinStepLine":31,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Fully Tailored Cybersecurity.\"","stepMatchArguments":[{"group":{"start":19,"value":"Fully Tailored Cybersecurity."}}]}]},
  {"pwTestLine":20,"pickleLine":34,"tags":["@journey:explore","@priority:nice-to-have","@kind:back-button"],"steps":[{"pwStepLine":21,"gherkinStepLine":35,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":22,"gherkinStepLine":36,"keywordType":"Action","textWithKeyword":"When I click the link to \"/cybersecurity\"","stepMatchArguments":[{"group":{"start":21,"value":"/cybersecurity"}}]},{"pwStepLine":23,"gherkinStepLine":37,"keywordType":"Action","textWithKeyword":"When I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":24,"gherkinStepLine":38,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]}]},
  {"pwTestLine":27,"pickleLine":41,"tags":["@journey:explore","@priority:nice-to-have","@kind:cross-journey"],"steps":[{"pwStepLine":28,"gherkinStepLine":42,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":29,"gherkinStepLine":43,"keywordType":"Action","textWithKeyword":"When I navigate directly to \"/\"","stepMatchArguments":[{"group":{"start":24,"value":"/"}}]},{"pwStepLine":30,"gherkinStepLine":44,"keywordType":"Action","textWithKeyword":"And I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":31,"gherkinStepLine":45,"keywordType":"Outcome","textWithKeyword":"Then no error message is shown in the form region","stepMatchArguments":[]}]},
  {"pwTestLine":34,"pickleLine":53,"tags":["@journey:explore","@priority:nice-to-have","@llm-composed","@kind:happy","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":35,"gherkinStepLine":54,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":36,"gherkinStepLine":55,"keywordType":"Action","textWithKeyword":"When I click the link to \"/cybersecurity\"","stepMatchArguments":[{"group":{"start":21,"value":"/cybersecurity"}}]},{"pwStepLine":37,"gherkinStepLine":56,"keywordType":"Outcome","textWithKeyword":"Then the page title contains \"Penetration Testing\"","stepMatchArguments":[{"group":{"start":25,"value":"Penetration Testing"}}]},{"pwStepLine":38,"gherkinStepLine":57,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Fully Tailored Cybersecurity.\"","stepMatchArguments":[{"group":{"start":24,"value":"Fully Tailored Cybersecurity."}}]}]},
  {"pwTestLine":41,"pickleLine":60,"tags":["@journey:explore","@priority:nice-to-have","@llm-composed","@kind:edge","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":42,"gherkinStepLine":61,"keywordType":"Context","textWithKeyword":"Given I open the page \"/cybersecurity\"","stepMatchArguments":[{"group":{"start":17,"value":"/cybersecurity"}}]},{"pwStepLine":43,"gherkinStepLine":62,"keywordType":"Outcome","textWithKeyword":"Then the page title contains \"Penetration Testing\"","stepMatchArguments":[{"group":{"start":25,"value":"Penetration Testing"}}]},{"pwStepLine":44,"gherkinStepLine":63,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Fully Tailored Cybersecurity.\"","stepMatchArguments":[{"group":{"start":24,"value":"Fully Tailored Cybersecurity."}}]}]},
  {"pwTestLine":47,"pickleLine":66,"tags":["@journey:explore","@priority:nice-to-have","@llm-composed","@kind:edge","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":48,"gherkinStepLine":67,"keywordType":"Context","textWithKeyword":"Given I open the page \"/cybersecurity\"","stepMatchArguments":[{"group":{"start":17,"value":"/cybersecurity"}}]},{"pwStepLine":49,"gherkinStepLine":68,"keywordType":"Action","textWithKeyword":"When I reload the page","stepMatchArguments":[]},{"pwStepLine":50,"gherkinStepLine":69,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Fully Tailored Cybersecurity.\"","stepMatchArguments":[{"group":{"start":24,"value":"Fully Tailored Cybersecurity."}}]}]},
]; // bdd-data-end