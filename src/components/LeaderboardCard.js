import React from "react";

import "../pages/Leaderboard.css";

// Components
import TeamDataContainer from "./TeamDataContainer";

export default function (props) {
  return (
    <div>
      <div className="leaderboardCard">
        <div className="team-container">
          <div className="team-icon">
            <img alt="" src={props.team.flag} />
          </div>
          <div className="team-text">{props.team.name_en} - Hakim</div>
         
        </div>
        <TeamDataContainer
            className="teamdata-container"
            stats={props.team}
          />
       </div>
    </div>
  );
}
