import React, { useState } from "react";
import { useForm } from "react-hook-form";
import { Navigate } from "react-router-dom";
import { Link } from "react-router-dom";
import "./Registration.css";

function Registration() {
  const [redirect, setRedirect] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { isValid, errors },
  } = useForm({
    mode: "onChange",
  });

  //Validates Password
  const validatePassword = (val) => {
    if (watch("password") != val) {
      return false;
    }
  };

  // register function
  const registerUser = async (data) => {
    try {
      const res = await fetch("http://localhost:8080/register", {
        method: "POST",
        body: JSON.stringify({
          full_name: data.name,
          email: data.email,
          password: data.password,
        }),
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
      });
    } catch (err) {}
  };

  // Sends data after form is complete
  const onSubmit = (data) => {
    registerUser(data).then(() => {
      setRedirect(true);
    });
  };

  if (redirect) return <Navigate to="/success" />;

  return (
    <div className="center-wrapper">
      <div className="reg-container">
        <div className="title">2022 World Cup Sweepstakes</div>
        <div className="content">
          {/* change action to redirect */}
          <form onSubmit={handleSubmit(onSubmit)}>
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
                {/* {errors?.password?.type === "required" && (
                  <p>This field is required</p>
                )} */}
              </div>
              {/* confirm password */}
              <div className="input-box">
                <span className="details">Confirm Password</span>
                <input
                  type="password"
                  placeholder="Confirm your password"
                  {...register("password_repeat", {
                    required: true,
                    validate: (val) => validatePassword(val),
                  })}
                />
                {/* {errors?.password_repeat?.type === "required" && (
                  <p>This field is required</p>
                )} */}
              </div>
            </div>
            {watch("password_repeat") !== watch("password") && (
              <p>password do not match</p>
            )}
            <div className="submitButton">
              <input type="submit" value="Register" disabled={!isValid} />
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export default Registration;
