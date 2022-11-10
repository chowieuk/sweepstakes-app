import React from "react";

import "./Leaderboardcard.css";

// Components
import TeamDataContainer from "./TeamDataContainer";

export default function (props) {
  return (
    <div>
      <div className="leadercard-container">
        <div className="team-container">
          <div className="team-icon">
            <img alt="" src={props.team.flag} />
          </div>
          <div className="team-text">{props.team.name_en} - Hakim</div>
          <TeamDataContainer
            className="teamdata-container"
            stats={props.team}
          />
        </div>
      </div>
    </div>
  );
}
