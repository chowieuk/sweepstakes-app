import React from "react";

import "./Leaderboardcard.css";

// Components
import LeaderboardCard from "./LeaderboardCard";

export default function LeaderboardWrapper(props) {
  props.data.map((group) => {
    return (
      <div key={group._id}>
        {group.teams.map((team) => {
          console.log(team);
          return <LeaderboardCard team={team} key={team.team_id} />;
        })}
      </div>
    );
  });
}
