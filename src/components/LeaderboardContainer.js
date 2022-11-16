import React from 'react'

import "../pages/Leaderboard.css";

//components
import LeaderboardHeader from './LeaderboardHeader'
import LeaderboardWrapper from './LeaderboardWrapper'

export default function LeaderboardContainer(props) {
  return (
    <div className="leaderboard-container">
      <LeaderboardHeader />
      <LeaderboardWrapper data={props.data}/>
    </div>
  )
}
