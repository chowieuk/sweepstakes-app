describe("pulls standing data from the api", () => {
  it("shows Qatar", () => {
    cy.visit("http://localhost:3000/leaderboard");

    cy.contains("Qatar").should("be.visible");
  });
  it("shows Qatar, Ecuador, Senegal and Nederland", () => {
    cy.visit("http://localhost:3000/leaderboard");

    cy.contains("Qatar").should("be.visible");
    cy.contains("Ecuador").should("be.visible");
    cy.contains("Senegal").should("be.visible");
    cy.contains("Nederland").should("be.visible");
  });
});

describe("pulls spec data data from the api", () => {
  it("shows MP not ID", () => {
    cy.visit("http://localhost:3000/leaderboard");

    cy.contains("MP").should("be.visible");
    cy.get("team_id").should("not.exist");
  });
});
