import React from "react";
import { Link } from "react-router-dom";
import { ReactComponent as YourSvg } from "./world-cup.svg";

import GoogleLogin from "../components/GoogleLogin";
import FacebookLogin from "../components/FacebookLogin";
import DevLogin from "../components/DevLogin";
import "./welcome.css";
export default function Welcome() {
  return (
    <>
      <div className="welcome-wrapper">
        <div className="welcome-container">
          <div className="welcome-container-row1">
            <div className="welcome-icon">
              <YourSvg className="welcome-icon" />
            </div>
            <div className="welcome-text">Welcome!</div>
          </div>
          <div className="welcome-container-row2">
            <div className="welcome-message">
              Chowie and Paddy invite you to join the 2022 worldcup sweepstakes
            </div>
          </div>
        </div>

        <div className="button-container">
          <Link
            to="/register"
            style={{ textDecoration: "none", color: "none" }}
          >
            <div className="button" action={"/registration"}>
              Register
            </div>
          </Link>
          <Link to="/login" style={{ textDecoration: "none", color: "none" }}>
            <div className="button" action={"/"}>
              login
            </div>
          </Link>
          <div className="social-login-container">
            <div className="social-buttons">
              <GoogleLogin/>
              <FacebookLogin/>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
