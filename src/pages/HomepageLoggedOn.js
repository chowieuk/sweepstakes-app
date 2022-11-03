import React from "react";
import { Link } from "react-router-dom";

import "./Registration.css";

//functions
import { getUserData } from "../functions/getUserData";


export default function HomePageLoggedOn() {


  // getUserData()
  return (
    <>
      <div className="post-reg-container">
        <div className="post-reg-container-row1">
          <div className="post-reg-text">Congratulations!</div>
        </div>
        <div className="post-reg-container-row2">
          <div className="post-reg-message">
           Your team is {}!
          </div>
        </div>
      </div>
    </>
  );
}
