import React from "react";


// Components
import LeaderboardCard from "./LeaderboardCard";

export default function LeaderboardWrapper(props) {
  return (
    <div className="leaderboard-wrapper">
        {props.data.map((group) => {

          return (
            <div key={group._id}>
              {group.teams.map((team) => {
                return <LeaderboardCard team={team} key={team.team_id} />;
              })}
            </div>
          );
        })}
    </div>
  )
}
