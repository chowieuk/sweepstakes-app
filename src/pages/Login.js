import React, { useState } from "react";
import { Navigate } from "react-router-dom";
import { Link } from "react-router-dom";
import { useForm } from "react-hook-form";

import './Login.css'

export default function Login() {
  const [redirect, setRedirect] = useState(false);


  const {
    handleSubmit,
    register,
    formState: { isValid, errors },
  } = useForm({
    mode: "onChange",
  });


//logon user
  const logonUser = async (data) => {
    try {
      const res = await fetch("http://localhost:8080/auth/mongo/login", {
        method: "POST",
        body: JSON.stringify({
          user: data.email,
          passwd: data.password,
        }),
        headers: {
          "Content-Type": "application/json",
          "Accept" : "application/json"
          // Consider adding "Authorization" : "some value" (https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization)
          // These two below are for responses, not requests.
          //"Access-Control-Allow-Origin": "*",
          //"Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept",
        },
      });
    } catch (err) {}
  };

  const onSubmit = (data) => {

    logonUser(data)
        .then(() => {
      setRedirect(true)
    })
  };

  if (redirect) return <Navigate to="/home" />;
 

  return (
    <div>
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="login-container">
          <div className="login-user-details">
                  {/* email */}
                <div className="login-input-box">
                  <span className="details">Email</span>
                  <input
                    type="email"
                    placeholder="Enter your email"
                    {...register("email", { required: true })}
                  />
                </div>
                {/* password */}
                <div className="login-input-box">
                  <span className="details">Password</span>
                  <input
                    type="password"
                    placeholder="Enter your password"
                    {...register("password", { required: true })}
                  />
                  {/* {errors?.password?.type === "required" && (
                    <p>This field is required</p>
                  )} */}
                </div>

                <div className="submitButton">
                  <input type="submit" value="Login" disabled={!isValid} />
                </div>

          </div>
        </div>
      </form>
    </div>
  )
}
