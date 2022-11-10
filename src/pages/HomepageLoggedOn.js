import React, { useState } from "react";
import { Link } from "react-router-dom";

import "./Registration.css";

export default function HomePageLoggedOn() {
  const [team, setTeam] = useState(false);
  const [teamFlag, setTeamFlag] = useState(false);

  const getUserData = async () => {
    try {
      const res = await fetch("http://localhost:8080/private_data", {
        //change endpoint as needed
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Headers":
            "Origin, X-Requested-With, Content-Type, Accept",
        },
      })
        .then((res) => res.json())
        .then((data) => {
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
            Welcome to the World Cup 2022 sweepstakes!
          </div>
        </div>
        <div className="post-reg-container-row2">
          <div className="post-reg-message">Your team is {team}!</div>
        </div>
      </div>
    </>
  );
}
