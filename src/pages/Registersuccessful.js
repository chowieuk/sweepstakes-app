import React from "react";
import { Link } from "react-router-dom";
import { ReactComponent as YourSvg } from "./world-cup.svg";

import "./Registration.css";

import "./Registration.css";
export default function RegistrationSuccessful() {
  return (
    <>
      <div className="post-reg-container">
        <div className="post-reg-container-row1">
          <div className="post-reg-text">Congratulations!</div>
        </div>
        <div className="post-reg-container-row2">
          <div className="post-reg-message">
            You are now registered to the 2022 World Cup sweepstakes!
            <br />
            Please LOGIN below to view your team
          </div>
          <Link
            className="button-container"
            to="/login"
            style={{ textDecoration: "none", color: "none" }}
          >
            <div className="button" action={"/"}>
              login
            </div>
          </Link>
        </div>
      </div>
    </>
  );
}
