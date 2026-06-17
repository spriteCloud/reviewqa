// Generated from: tests/e2e/features/spritecloud-com-explore-contact.feature
import { test } from "playwright-bdd";

test.describe('WwwSpritecloudCom — explore journey', () => {

  test('explore journey reaches its terminal page', { tag: ['@journey:explore', '@priority:nice-to-have', '@smoke'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await And('the page title contains "spriteCloud - Test your software, not your reputation!"', null, { page }); 
    await And('the main heading reads "Test your software, not your reputation."', null, { page }); 
    await When('I click the link to "/contact"', null, { page }); 
    await Then('I see the heading "Let\'s Chat"', null, { page }); 
    await And('the page title contains "spriteCloud — Meeting Booking Form"', null, { page }); 
  });

  test('explore — deep-link to the terminal page renders correctly', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:resume'] }, async ({ Given, Then, page }) => { 
    await Given('I open the page "/contact"', null, { page }); 
    await Then('I see the heading "Let\'s Chat"', null, { page }); 
  });

  test('explore — back button after navigation returns to landing', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:back-button'] }, async ({ Given, When, Then, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I click the link to "/contact"', null, { page }); 
    await When('I go back in the browser history', null, { page }); 
    await Then('the main heading reads "Test your software, not your reputation."', null, { page }); 
  });

  test('explore — switching to landing and back leaves no broken state', { tag: ['@journey:explore', '@priority:nice-to-have', '@kind:cross-journey'] }, async ({ Given, When, Then, And, page }) => { 
    await Given('I open the landing page', null, { page }); 
    await When('I navigate directly to "/"', null, { page }); 
    await And('I go back in the browser history', null, { page }); 
    await Then('no error message is shown in the form region', null, { page }); 
  });

});

// == technical section ==

test.use({
  $test: [({}, use) => use(test), { scope: 'test', box: true }],
  $uri: [({}, use) => use('tests/e2e/features/spritecloud-com-explore-contact.feature'), { scope: 'test', box: true }],
  $bddFileData: [({}, use) => use(bddFileData), { scope: "test", box: true }],
});

const bddFileData = [ // bdd-data-start
  {"pwTestLine":6,"pickleLine":20,"tags":["@journey:explore","@priority:nice-to-have","@smoke"],"steps":[{"pwStepLine":7,"gherkinStepLine":21,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":8,"gherkinStepLine":22,"keywordType":"Context","textWithKeyword":"And the page title contains \"spriteCloud - Test your software, not your reputation!\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud - Test your software, not your reputation!"}}]},{"pwStepLine":9,"gherkinStepLine":23,"keywordType":"Context","textWithKeyword":"And the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]},{"pwStepLine":10,"gherkinStepLine":24,"keywordType":"Action","textWithKeyword":"When I click the link to \"/contact\"","stepMatchArguments":[{"group":{"start":21,"value":"/contact"}}]},{"pwStepLine":11,"gherkinStepLine":25,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Let's Chat\"","stepMatchArguments":[{"group":{"start":19,"value":"Let's Chat"}}]},{"pwStepLine":12,"gherkinStepLine":26,"keywordType":"Outcome","textWithKeyword":"And the page title contains \"spriteCloud — Meeting Booking Form\"","stepMatchArguments":[{"group":{"start":25,"value":"spriteCloud — Meeting Booking Form"}}]}]},
  {"pwTestLine":15,"pickleLine":29,"tags":["@journey:explore","@priority:nice-to-have","@kind:resume"],"steps":[{"pwStepLine":16,"gherkinStepLine":30,"keywordType":"Context","textWithKeyword":"Given I open the page \"/contact\"","stepMatchArguments":[{"group":{"start":17,"value":"/contact"}}]},{"pwStepLine":17,"gherkinStepLine":31,"keywordType":"Outcome","textWithKeyword":"Then I see the heading \"Let's Chat\"","stepMatchArguments":[{"group":{"start":19,"value":"Let's Chat"}}]}]},
  {"pwTestLine":20,"pickleLine":34,"tags":["@journey:explore","@priority:nice-to-have","@kind:back-button"],"steps":[{"pwStepLine":21,"gherkinStepLine":35,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":22,"gherkinStepLine":36,"keywordType":"Action","textWithKeyword":"When I click the link to \"/contact\"","stepMatchArguments":[{"group":{"start":21,"value":"/contact"}}]},{"pwStepLine":23,"gherkinStepLine":37,"keywordType":"Action","textWithKeyword":"When I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":24,"gherkinStepLine":38,"keywordType":"Outcome","textWithKeyword":"Then the main heading reads \"Test your software, not your reputation.\"","stepMatchArguments":[{"group":{"start":24,"value":"Test your software, not your reputation."}}]}]},
  {"pwTestLine":27,"pickleLine":41,"tags":["@journey:explore","@priority:nice-to-have","@kind:cross-journey"],"steps":[{"pwStepLine":28,"gherkinStepLine":42,"keywordType":"Context","textWithKeyword":"Given I open the landing page","stepMatchArguments":[]},{"pwStepLine":29,"gherkinStepLine":43,"keywordType":"Action","textWithKeyword":"When I navigate directly to \"/\"","stepMatchArguments":[{"group":{"start":24,"value":"/"}}]},{"pwStepLine":30,"gherkinStepLine":44,"keywordType":"Action","textWithKeyword":"And I go back in the browser history","stepMatchArguments":[]},{"pwStepLine":31,"gherkinStepLine":45,"keywordType":"Outcome","textWithKeyword":"Then no error message is shown in the form region","stepMatchArguments":[]}]},
]; // bdd-data-end