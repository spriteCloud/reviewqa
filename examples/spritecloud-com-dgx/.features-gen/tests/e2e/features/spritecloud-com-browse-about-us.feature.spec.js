// Generated from: tests/e2e/features/spritecloud-com-browse-about-us.feature
import { test } from "playwright-bdd";

test.describe('WwwSpritecloudCom — browse journey', () => {

  test('browse journey reaches its terminal page', { tag: ['@journey:browse', '@priority:standard', '@smoke'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await And('the page title contains "spriteCloud - Test your software, not your reputation!"', null, { page }); 
    await And('the main heading reads "Test your software, not your reputation."', null, { page }); 
    await When('I click the link to "/guides"', null, { page }); 
    await And('the page title contains "spriteCloud - Your Software Testing and QA partner"', null, { page }); 
    await When('I click the link to "/about-us"', null, { page }); 
    await Then('I see the heading "Testing is in our DNA."', null, { page }); 
    await And('the page title contains "About Us"', null, { page }); 
  });

  test('browse — deep-link to the terminal page renders correctly', { tag: ['@journey:browse', '@priority:standard', '@kind:resume'] }, async ({ Given, Then, page }) => { 
    await Given('I open the page "/about-us"', null, { page }); 
    await Then('I see the heading "Testing is in our DNA."', null, { page }); 
  });

  test('browse — back button after navigation returns to landing', { tag: ['@journey:browse', '@priority:standard', '@kind:back-button'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/about-us"', null, { page }); 
    await When('I go back in the browser history', null, { page }); 
    await Then('the main heading reads "Test your software, not your reputation."', null, { page }); 
  });

  test('browse — switching to landing and back leaves no broken state', { tag: ['@journey:browse', '@priority:standard', '@kind:cross-journey'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I navigate directly to "/"', null, { page }); 
    await And('I go back in the browser history', null, { page }); 
    await Then('no error message is shown in the form region', null, { page }); 
  });

});

// == technical section ==

test.use({
  $test: [({}, use) => use(test), { scope: 'test', box: true }],
  $uri: [({}, use) => use('tests/e2e/features/spritecloud-com-browse-about-us.feature'), { scope: 'test', box: true }],
  $bddFileData: [({}, use) => use(bddFileData), { scope: "test", box: true }],
});

const bddFileData = [ // bdd-data-start
  {"pwTestLine":6,"pickleLine":20,"tags":["@journey:browse","@priority:standard","@smoke"],"steps":[{"pwStepLine":7,"gherkinStepLine":21,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":8,"gherkinStepLine":22,"keywordType":"Context","textWithKeyword":"And the page title contains \"spriteCloud - Test your software, not your reputation!\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Test your software, not your reputation!"}}]},{"pwStepLine":9,"gherkinStepLine":23,"keywordType":"Context","textWithKeyword":"And the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]},{"pwStepLine":10,"gherkinStepLine":24,"keywordType":"Action","textWithKeyword":"When I click the link to \"/guides\"","stepMatchArguments":[{"group":{"start":21,"value":"/guides"}}]},{"pwStepLine":11,"gherkinStepLine":25,"keywordType":"Action","textWithKeyword":"And the page title contains \"spriteCloud - Your Software Testing and QA partner\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Your Software Testing and QA partner"}}]},{"pwStepLine":12,"gherkinStepLine":26,"keywordType":"Action","textWithKeyword":"When I click the link to \"/about-us\"","stepMatchArguments":[{"group":{"start":21,"value":"/about-us"}}]},{"pwStepLine":13,"gherkinStepLine":27,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Testing is in our DNA.\"","stepMatchArguments":[{"group":{"start":19,"value":"Testing is in our DNA."}}]},{"pwStepLine":14,"gherkinStepLine":28,"keywordType":"Outcome","textWithKeyword":"And the page title contains \"About Us\"","stepMatchArguments":[{"group":{"start":25,"value":"About Us"}}]}]},
  {"pwTestLine":17,"pickleLine":31,"tags":["@journey:browse","@priority:standard","@kind:resume"],"steps":[{"pwStepLine":18,"gherkinStepLine":32,"keywordType":"Context","textWithKeyword":"Given I open the page \"/about-us\"","stepMatchArguments":[{"group":{"start":17,"value":"/about-us"}}]},{"pwStepLine":19,"gherkinStepLine":33,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Testing is in our DNA.\"","stepMatchArguments":[{"group":{"start":19,"value":"Testing is in our DNA."}}]}]},
  {"pwTestLine":22,"pickleLine":36,"tags":["@journey:browse","@priority:standard","@kind:back-button"],"steps":[{"pwStepLine":23,"gherkinStepLine":37,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":24,"gherkinStepLine":38,"keywordType":"Action","textWithKeyword":"When I click the link to \"/about-us\"","stepMatchArguments":[{"group":{"start":21,"value":"/about-us"}}]},{"pwStepLine":25,"gherkinStepLine":39,"keywordType":"Action","textWithKeyword":"When I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":26,"gherkinStepLine":40,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]}]},
  {"pwTestLine":29,"pickleLine":43,"tags":["@journey:browse","@priority:standard","@kind:cross-journey"],"steps":[{"pwStepLine":30,"gherkinStepLine":44,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":31,"gherkinStepLine":45,"keywordType":"Action","textWithKeyword":"When I navigate directly to \"/\"","stepMatchArguments":[{"group":{"start":24,"value":"/"}}]},{"pwStepLine":32,"gherkinStepLine":46,"keywordType":"Action","textWithKeyword":"And I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":33,"gherkinStepLine":47,"keywordType":"Outcome","textWithKeyword":"Then no error message is shown in the form region","stepMatchArguments":[]}]},
]; // bdd-data-end