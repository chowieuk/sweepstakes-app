import React from "react";

export default function MatchCardContainer({ props }) {
  props.matchData.map((match) => {
    return (
      <div className="container-fluid" key={match._id}>
        Group: {match.group} - Matchday {match.matchday} of 3<br />
        {match.home_team_en} vs {match.away_team_en} <br />
        {match.local_date}
      </div>
    );
  });
}
