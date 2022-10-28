import React from 'react'
import { useForm } from "react-hook-form";

export default function Login() {



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
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Headers":
            "Origin, X-Requested-With, Content-Type, Accept",
        },
      });
    } catch (err) {}
  };

  const onSubmit = (data) => {
    logonUser(data);
  };


  return (
    <div>
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="user-details">
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

            <div className="submitButton">
              <input type="submit" value="Register" disabled={!isValid} />
            </div>
        </div>
      </form>
    </div>
  )
}
