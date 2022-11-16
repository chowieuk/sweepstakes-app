import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";

import InfiniteLooper from "../components/InfiniteLooper.tsx";
import Flags from "../components/Flags.tsx";

import "./Registration.css";

export default function HomePageLoggedOn() {
  const [flags, setFlags] = useState([
    "https://upload.wikimedia.org/wikipedia/commons/b/be/Flag_of_England.svg",
    "https://upload.wikimedia.org/wikipedia/commons/a/a4/Flag_of_the_United_States.svg",
    "https://upload.wikimedia.org/wikipedia/commons/e/e8/Flag_of_Ecuador.svg",
    "https://upload.wikimedia.org/wikipedia/commons/f/fd/Flag_of_Senegal.svg",
  ]);

  const [team, setTeam] = useState(false);
  const [teamFlag, setTeamFlag] = useState(
    "https://upload.wikimedia.org/wikipedia/commons/c/ca/Flag_of_Iran.svg"
  );
  const [userName, setUserName] = useState(false);

  const getUserData = async () => {
    try {
      const res = await fetch("http://localhost:8080/private_data", {
        //change endpoint as needed
        method: "GET",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
      })
        .then((res) => res.json())
        .then((data) => {
          setUserName(data.userInfo.name.split(" ")[0]);
          setTeam(data.userInfo.attrs.team_name);
          setTeamFlag(data.userInfo.attrs.team_flag);
        });
    } catch (err) {}
  };

  useEffect(() => {
    getUserData();
  }, []);

  useEffect(() => {
    setFlags([teamFlag, ...flags]);
  }, [teamFlag]);

  return (
    <>
      <div className="post-reg-container">
        <div className="post-reg-container-row1">
          <div className="post-reg-text">Good luck {userName}!</div>
        </div>
        <div className="post-reg-container-row2">
          <div className="post-reg-message">
            <div className="smallViewport">
              <InfiniteLooper
                speed={1.5}
                direction="right"
                animState="true"
                userTeam={team}
              >
                <Flags flags={flags} />
              </InfiniteLooper>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
