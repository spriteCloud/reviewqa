// Generated from: tests/e2e/features/spritecloud-com-research-case-study-ecomm-platform.feature
import { test } from "playwright-bdd";

test.describe('WwwSpritecloudCom — research journey', () => {

  test('research journey reaches its terminal page', { tag: ['@journey:research', '@priority:standard', '@smoke'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await And('the page title contains "spriteCloud - Test your software, not your reputation!"', null, { page }); 
    await And('the main heading reads "Test your software, not your reputation."', null, { page }); 
    await When('I click the link to "/case-study-ecomm-platform"', null, { page }); 
    await Then('I see the heading "Performance Testing for an eCommerce Platform"', null, { page }); 
    await And('the page title contains "Case Study - eCommerce Platform"', null, { page }); 
  });

  test('research — deep-link to the terminal page renders correctly', { tag: ['@journey:research', '@priority:standard', '@kind:resume'] }, async ({ Given, Then, page }) => { 
    await Given('I open the page "/case-study-ecomm-platform"', null, { page }); 
    await Then('I see the heading "Performance Testing for an eCommerce Platform"', null, { page }); 
  });

  test('research — back button after navigation returns to landing', { tag: ['@journey:research', '@priority:standard', '@kind:back-button'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/case-study-ecomm-platform"', null, { page }); 
    await When('I go back in the browser history', null, { page }); 
    await Then('the main heading reads "Test your software, not your reputation."', null, { page }); 
  });

  test('reload during form submission', { tag: ['@journey:research', '@priority:standard', '@llm-composed', '@kind:edge', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/case-study-ecomm-platform"', null, { page }); 
    await When('I reload the page', null, { page }); 
    await Then('the page title contains "Case Study - eCommerce Platform"', null, { page }); 
  });

  test('open destination directly', { tag: ['@journey:research', '@priority:standard', '@llm-composed', '@kind:variant', '@model:qwen3-coder-next-latest'] }, async ({ When, Then, page }) => { 
    await When('I navigate directly to "/case-study-ecomm-platform"', null, { page }); 
    await Then('the page title contains "Case Study - eCommerce Platform"', null, { page }); 
    await Then('the main heading reads "Performance Testing for an eCommerce Platform"', null, { page }); 
  });

  test('landing page reloads to self', { tag: ['@journey:research', '@priority:standard', '@llm-composed', '@kind:edge', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I am on the landing page', null, { page }); 
    await When('I reload the page', null, { page }); 
    await Then('the main heading reads "Test your software, not your reputation."', null, { page }); 
  });

  test('submit form then reload without navigation', { tag: ['@journey:research', '@priority:standard', '@llm-composed', '@kind:edge', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I submit the form without filling any required field', null, { page }); 
    await When('I reload the page', null, { page }); 
    await Then('no error message is shown in the form region', null, { page }); 
  });

});

// == technical section ==

test.use({
  $test: [({}, use) => use(test), { scope: 'test', box: true }],
  $uri: [({}, use) => use('tests/e2e/features/spritecloud-com-research-case-study-ecomm-platform.feature'), { scope: 'test', box: true }],
  $bddFileData: [({}, use) => use(bddFileData), { scope: "test", box: true }],
});

const bddFileData = [ // bdd-data-start
  {"pwTestLine":6,"pickleLine":20,"tags":["@journey:research","@priority:standard","@smoke"],"steps":[{"pwStepLine":7,"gherkinStepLine":21,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":8,"gherkinStepLine":22,"keywordType":"Context","textWithKeyword":"And the page title contains \"spriteCloud - Test your software, not your reputation!\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Test your software, not your reputation!"}}]},{"pwStepLine":9,"gherkinStepLine":23,"keywordType":"Context","textWithKeyword":"And the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]},{"pwStepLine":10,"gherkinStepLine":24,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-study-ecomm-platform\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-study-ecomm-platform"}}]},{"pwStepLine":11,"gherkinStepLine":25,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Performance Testing for an eCommerce Platform\"","stepMatchArguments":[{"group":{"start":19,"value":"Performance Testing for an eCommerce Platform"}}]},{"pwStepLine":12,"gherkinStepLine":26,"keywordType":"Outcome","textWithKeyword":"And the page title contains \"Case Study - eCommerce Platform\"","stepMatchArguments":[{"group":{"start":25,"value":"Case Study - eCommerce Platform"}}]}]},
  {"pwTestLine":15,"pickleLine":29,"tags":["@journey:research","@priority:standard","@kind:resume"],"steps":[{"pwStepLine":16,"gherkinStepLine":30,"keywordType":"Context","textWithKeyword":"Given I open the page \"/case-study-ecomm-platform\"","stepMatchArguments":[{"group":{"start":17,"value":"/case-study-ecomm-platform"}}]},{"pwStepLine":17,"gherkinStepLine":31,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Performance Testing for an eCommerce Platform\"","stepMatchArguments":[{"group":{"start":19,"value":"Performance Testing for an eCommerce Platform"}}]}]},
  {"pwTestLine":20,"pickleLine":34,"tags":["@journey:research","@priority:standard","@kind:back-button"],"steps":[{"pwStepLine":21,"gherkinStepLine":35,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":22,"gherkinStepLine":36,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-study-ecomm-platform\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-study-ecomm-platform"}}]},{"pwStepLine":23,"gherkinStepLine":37,"keywordType":"Action","textWithKeyword":"When I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":24,"gherkinStepLine":38,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]}]},
  {"pwTestLine":27,"pickleLine":46,"tags":["@journey:research","@priority:standard","@llm-composed","@kind:edge","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":28,"gherkinStepLine":47,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":29,"gherkinStepLine":48,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-study-ecomm-platform\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-study-ecomm-platform"}}]},{"pwStepLine":30,"gherkinStepLine":49,"keywordType":"Action","textWithKeyword":"When I reload the page","stepMatchArguments":[]},{"pwStepLine":31,"gherkinStepLine":50,"keywordType":"Outcome","textWithKeyword":"Then the page title contains \"Case Study - eCommerce Platform\"","stepMatchArguments":[{"group":{"start":25,"value":"Case Study - eCommerce Platform"}}]}]},
  {"pwTestLine":34,"pickleLine":53,"tags":["@journey:research","@priority:standard","@llm-composed","@kind:variant","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":35,"gherkinStepLine":54,"keywordType":"Action","textWithKeyword":"When I navigate directly to \"/case-study-ecomm-platform\"","stepMatchArguments":[{"group":{"start":24,"value":"/case-study-ecomm-platform"}}]},{"pwStepLine":36,"gherkinStepLine":55,"keywordType":"Outcome","textWithKeyword":"Then the page title contains \"Case Study - eCommerce Platform\"","stepMatchArguments":[{"group":{"start":25,"value":"Case Study - eCommerce Platform"}}]},{"pwStepLine":37,"gherkinStepLine":56,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Performance Testing for an eCommerce Platform\"","stepMatchArguments":[{"group":{"start":24,"value":"Performance Testing for an eCommerce Platform"}}]}]},
  {"pwTestLine":40,"pickleLine":59,"tags":["@journey:research","@priority:standard","@llm-composed","@kind:edge","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":41,"gherkinStepLine":60,"keywordType":"Context","textWithKeyword":"Given I am on the landing page","stepMatchArguments":[]},{"pwStepLine":42,"gherkinStepLine":61,"keywordType":"Action","textWithKeyword":"When I reload the page","stepMatchArguments":[]},{"pwStepLine":43,"gherkinStepLine":62,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]}]},
  {"pwTestLine":46,"pickleLine":65,"tags":["@journey:research","@priority:standard","@llm-composed","@kind:edge","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":47,"gherkinStepLine":66,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":48,"gherkinStepLine":67,"keywordType":"Action","textWithKeyword":"When I submit the form without filling any required field","stepMatchArguments":[]},{"pwStepLine":49,"gherkinStepLine":68,"keywordType":"Action","textWithKeyword":"When I reload the page","stepMatchArguments":[]},{"pwStepLine":50,"gherkinStepLine":69,"keywordType":"Outcome","textWithKeyword":"Then no error message is shown in the form region","stepMatchArguments":[]}]},
]; // bdd-data-end