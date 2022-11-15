import React from "react";
import { Link } from "react-router-dom";
import { ReactComponent as YourSvg } from "./world-cup.svg";

import "./Registration.css";

import "./Registration.css";
export default function RegistrationFail() {
  return (
    <>
      <div className="post-reg-container">
        <div className="post-reg-container-row1">
          <div className="post-reg-text">Whoops!</div>
        </div>
        <div className="post-reg-container-row2">
          <div className="post-reg-message">
            I'm sorry you were too late to register!
            <br />
            You are on our waitlist and will be alerted if anything becomes
            availible.
          </div>
        </div>
      </div>
    </>
  );
}
