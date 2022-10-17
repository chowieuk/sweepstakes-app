// team picker

const teamPicker = (teams) => {
  const availibleTeams = teams.filter((team) => team.availible);

  const randomTeam =
    availibleTeams[Math.floor(Math.random() * availibleTeams.length)];

  return randomTeam;
};

module.exports = teamPicker;
