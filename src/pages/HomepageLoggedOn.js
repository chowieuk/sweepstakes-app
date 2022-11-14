import React, { useState } from "react";
import { Link } from "react-router-dom";

import "./Registration.css";

export default function HomePageLoggedOn() {
  const [team, setTeam] = useState(false);
  const [teamFlag, setTeamFlag] = useState(false);
  const [userName, setUserName] = useState(false);

  const getUserData = async () => {
    try {
      const res = await fetch("http://localhost:8080/private_data", {
        //change endpoint as needed
        method: "GET",
        credentials: 'include',
        headers: {
          "Content-Type": "application/json",
          // "Access-Control-Allow-Origin": "*",
          // "Access-Control-Allow-Headers":
          //   "Origin, X-Requested-With, Content-Type, Accept",
        },
      })
        .then((res) => res.json())
        .then((data) => {
          setUserName(data.userInfo.name);
          setTeam(data.userInfo.attrs.team_name);
          setTeamFlag(data.userInfo.attrs.team_flag);
        });
    } catch (err) {}
  };
  getUserData();

  return (
    <>
      <div className="post-reg-container">
        <div className="post-reg-container-row1">
          <div className="post-reg-text">
            Congratulations {userName}
          </div>
        </div>
        <div className="post-reg-container-row2">
          <div className="post-reg-message">
            Your team is {team}!
            <img className="flag-image" src={teamFlag}/>
          </div>
          
        </div>
      </div>
    </>
  );
}
