import { useState, useEffect, useRef } from "react";
import ReactDOM from "react-dom";
import { useForm } from "react-hook-form";
import "./App.css";
// import { db } from "./firebase-config";
// import {
//   collection,
//   getDocs,
//   addDoc,
//   updateDoc,
//   deleteDoc,
//   doc,
// } from "firebase/firestore";

//dummyData
const teams = [
  {
    name: "Qatar",
    availible: true,
  },
  {
    name: "Ecuador",
    availible: true,
  },
  {
    name: "Senegal",
    availible: false,
  },
  {
    name: "Netherlands",
    availible: true,
  },
];
// team picker
const availibleTeams = teams.filter((team) => team.availible);

const randomTeam =
  availibleTeams[Math.floor(Math.random() * availibleTeams.length)];
console.log(randomTeam);

function App() {
  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isValid },
  } = useForm({
    mode: "onChange",
  });

  const onSubmit = (data) => {
    console.log(data);
  }; // your form submit function which will invoke after successful validation

  return (
    <>
      <div className="container">
        <div className="title">2022 World Cup Sweepstakes</div>
        <div className="content">
          {/* change action to redirect */}
          <form
            onSubmit={(e) => {
              handleSubmit(onSubmit)(e)
                // you will have to catch those error and handle them
                .catch(() => {});
            }}
          >
            <div className="user-details">
              {/* name */}
              <div className="input-box">
                <span className="details">Full Name</span>
                <input
                  type="text"
                  placeholder="Enter your name"
                  {...register("name", { required: true })}
                />
              </div>
              {/* email */}
              <div className="input-box">
                <span className="details">Email</span>
                <input
                  type="email"
                  placeholder="Enter your email"
                  {...register("email", { required: true })}
                />
              </div>
              {/* password */}
              <div className="input-box">
                <span className="details">Password</span>
                <input
                  type="password"
                  placeholder="Enter your password"
                  {...register("password", { required: true })}
                />
                {errors?.password?.type === "required" && (
                  <p>This field is required</p>
                )}
              </div>
              {/* confirm password */}
              <div className="input-box">
                <span className="details">Confirm Password</span>
                <input
                  type="password"
                  placeholder="Confirm your password"
                  {...register("password_repeat", {
                    required: true,
                    validate: (val) => {
                      if (watch("password") != val) {
                        return false;
                      }
                    },
                  })}
                />
                {errors?.password_repeat?.type === "required" && (
                  <p>This field is required</p>
                )}
              </div>
            </div>
            {watch("password_repeat") !== watch("password") && (
              <p>password do not match</p>
            )}
            <div className="button">
              <input type="submit" value="Register" hidden={!isValid} />
            </div>
          </form>
        </div>
      </div>
    </>
  );
}

export default App;
