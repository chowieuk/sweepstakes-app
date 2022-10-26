import { useForm } from "react-hook-form";
import "../App.css";
import "./Registration.css";

function Registration() {
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

  // Sends data after form is complete
  const onSubmit = (data) => {
    console.log(data);
  };

  return (
    <>
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
    </>
  );
}

export default Registration;
