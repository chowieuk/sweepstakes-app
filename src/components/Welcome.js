import React from "react";
import { Link } from "react-router-dom";

export default function Welcome() {
  return (
    <>
      <section className="welcomeBackground">
        <div className="welcomeContainer">
          Chowie and Paddy welcome you to participate in the 2022 world cup
          sweepstakes!
        </div>
        <div className="buttonContainer">
          <Link
            to="/registration"
            style={{ textDecoration: "none", color: "none" }}
          >
            <div className="button" action={"/registration"}>
              Register
            </div>
          </Link>
          <Link
            to="/registration"
            style={{ textDecoration: "none", color: "none" }}
          >
            <div className="button" action={"/"}>
              login
            </div>
          </Link>
        </div>
      </section>
    </>
  );
}
