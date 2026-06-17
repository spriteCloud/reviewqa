// Generated from: tests/e2e/features/spritecloud-com-research-case-study-grandvision.feature
import { test } from "playwright-bdd";

test.describe('WwwSpritecloudCom — research journey', () => {

  test('research journey reaches its terminal page', { tag: ['@journey:research', '@priority:standard', '@smoke'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await And('the page title contains "spriteCloud - Test your software, not your reputation!"', null, { page }); 
    await And('the main heading reads "Test your software, not your reputation."', null, { page }); 
    await When('I navigate directly to "/case-study-grandvision"', null, { page }); 
    await Then('I see the heading "GrandVision"', null, { page }); 
    await And('the page title contains "Case Study - GrandVision"', null, { page }); 
  });

  test('research — deep-link to the terminal page renders correctly', { tag: ['@journey:research', '@priority:standard', '@kind:resume'] }, async ({ Given, Then, page }) => { 
    await Given('I open the page "/case-study-grandvision"', null, { page }); 
    await Then('I see the heading "GrandVision"', null, { page }); 
  });

  test('research — back button after navigation returns to landing', { tag: ['@journey:research', '@priority:standard', '@kind:back-button'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/case-study-grandvision"', null, { page }); 
    await When('I go back in the browser history', null, { page }); 
    await Then('the main heading reads "Test your software, not your reputation."', null, { page }); 
  });

  test('Navigate to case study via top link', { tag: ['@journey:research', '@priority:standard', '@llm-composed', '@kind:edge', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/case-studies"', null, { page }); 
    await When('I click the link to "/case-study-grandvision"', null, { page }); 
    await Then('the page title contains "Case Study - GrandVision"', null, { page }); 
    await Then('the main heading reads "GrandVision"', null, { page }); 
  });

  test('Reload case study page mid-flow', { tag: ['@journey:research', '@priority:standard', '@llm-composed', '@kind:variant', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/case-study-grandvision"', null, { page }); 
    await When('I reload the page', null, { page }); 
    await Then('the page title contains "Case Study - GrandVision"', null, { page }); 
    await Then('the main heading reads "GrandVision"', null, { page }); 
  });

  test('Direct navigation to case study', { tag: ['@journey:research', '@priority:standard', '@llm-composed', '@model:qwen3-coder-next-latest'] }, async ({ Given, When, Then, context, page }) => { 
    await Given('I am not signed in', null, { context }); 
    await When('I navigate directly to "/case-study-grandvision"', null, { page }); 
    await Then('the page title contains "Case Study - GrandVision"', null, { page }); 
    await Then('the main heading reads "GrandVision"', null, { page }); 
  });

});

// == technical section ==

test.use({
  $test: [({}, use) => use(test), { scope: 'test', box: true }],
  $uri: [({}, use) => use('tests/e2e/features/spritecloud-com-research-case-study-grandvision.feature'), { scope: 'test', box: true }],
  $bddFileData: [({}, use) => use(bddFileData), { scope: "test", box: true }],
});

const bddFileData = [ // bdd-data-start
  {"pwTestLine":6,"pickleLine":20,"tags":["@journey:research","@priority:standard","@smoke"],"steps":[{"pwStepLine":7,"gherkinStepLine":21,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":8,"gherkinStepLine":22,"keywordType":"Context","textWithKeyword":"And the page title contains \"spriteCloud - Test your software, not your reputation!\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Test your software, not your reputation!"}}]},{"pwStepLine":9,"gherkinStepLine":23,"keywordType":"Context","textWithKeyword":"And the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]},{"pwStepLine":10,"gherkinStepLine":24,"keywordType":"Action","textWithKeyword":"When I navigate directly to \"/case-study-grandvision\"","stepMatchArguments":[{"group":{"start":24,"value":"/case-study-grandvision"}}]},{"pwStepLine":11,"gherkinStepLine":25,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"GrandVision\"","stepMatchArguments":[{"group":{"start":19,"value":"GrandVision"}}]},{"pwStepLine":12,"gherkinStepLine":26,"keywordType":"Outcome","textWithKeyword":"And the page title contains \"Case Study - GrandVision\"","stepMatchArguments":[{"group":{"start":25,"value":"Case Study - GrandVision"}}]}]},
  {"pwTestLine":15,"pickleLine":29,"tags":["@journey:research","@priority:standard","@kind:resume"],"steps":[{"pwStepLine":16,"gherkinStepLine":30,"keywordType":"Context","textWithKeyword":"Given I open the page \"/case-study-grandvision\"","stepMatchArguments":[{"group":{"start":17,"value":"/case-study-grandvision"}}]},{"pwStepLine":17,"gherkinStepLine":31,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"GrandVision\"","stepMatchArguments":[{"group":{"start":19,"value":"GrandVision"}}]}]},
  {"pwTestLine":20,"pickleLine":34,"tags":["@journey:research","@priority:standard","@kind:back-button"],"steps":[{"pwStepLine":21,"gherkinStepLine":35,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":22,"gherkinStepLine":36,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-study-grandvision\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-study-grandvision"}}]},{"pwStepLine":23,"gherkinStepLine":37,"keywordType":"Action","textWithKeyword":"When I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":24,"gherkinStepLine":38,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]}]},
  {"pwTestLine":27,"pickleLine":46,"tags":["@journey:research","@priority:standard","@llm-composed","@kind:edge","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":28,"gherkinStepLine":47,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":29,"gherkinStepLine":48,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-studies\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-studies"}}]},{"pwStepLine":30,"gherkinStepLine":49,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-study-grandvision\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-study-grandvision"}}]},{"pwStepLine":31,"gherkinStepLine":50,"keywordType":"Outcome","textWithKeyword":"Then the page title contains \"Case Study - GrandVision\"","stepMatchArguments":[{"group":{"start":25,"value":"Case Study - GrandVision"}}]},{"pwStepLine":32,"gherkinStepLine":51,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"GrandVision\"","stepMatchArguments":[{"group":{"start":24,"value":"GrandVision"}}]}]},
  {"pwTestLine":35,"pickleLine":54,"tags":["@journey:research","@priority:standard","@llm-composed","@kind:variant","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":36,"gherkinStepLine":55,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":37,"gherkinStepLine":56,"keywordType":"Action","textWithKeyword":"When I click the link to \"/case-study-grandvision\"","stepMatchArguments":[{"group":{"start":21,"value":"/case-study-grandvision"}}]},{"pwStepLine":38,"gherkinStepLine":57,"keywordType":"Action","textWithKeyword":"When I reload the page","stepMatchArguments":[]},{"pwStepLine":39,"gherkinStepLine":58,"keywordType":"Outcome","textWithKeyword":"Then the page title contains \"Case Study - GrandVision\"","stepMatchArguments":[{"group":{"start":25,"value":"Case Study - GrandVision"}}]},{"pwStepLine":40,"gherkinStepLine":59,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"GrandVision\"","stepMatchArguments":[{"group":{"start":24,"value":"GrandVision"}}]}]},
  {"pwTestLine":43,"pickleLine":62,"tags":["@journey:research","@priority:standard","@llm-composed","@model:qwen3-coder-next-latest"],"steps":[{"pwStepLine":44,"gherkinStepLine":63,"keywordType":"Context","textWithKeyword":"Given I am not signed in","stepMatchArguments":[]},{"pwStepLine":45,"gherkinStepLine":64,"keywordType":"Action","textWithKeyword":"When I navigate directly to \"/case-study-grandvision\"","stepMatchArguments":[{"group":{"start":24,"value":"/case-study-grandvision"}}]},{"pwStepLine":46,"gherkinStepLine":65,"keywordType":"Outcome","textWithKeyword":"Then the page title contains \"Case Study - GrandVision\"","stepMatchArguments":[{"group":{"start":25,"value":"Case Study - GrandVision"}}]},{"pwStepLine":47,"gherkinStepLine":66,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"GrandVision\"","stepMatchArguments":[{"group":{"start":24,"value":"GrandVision"}}]}]},
]; // bdd-data-end