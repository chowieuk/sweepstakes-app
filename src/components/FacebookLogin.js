import React from "react";
import { useRef, useEffect } from "react";
import { socialLogin, socialLoginErrorHandler } from "../functions/socialLogin";

const FacebookLogin = () => {
  const ref = useRef(null);

  useEffect(() => {
    const handleClick = (e) => {
      e.preventDefault();
      socialLogin("facebook")
        .then(() => {
          window.location.replace(window.location.href + "home");
        })
        .catch(socialLoginErrorHandler);
    };

    const element = ref.current;

    element.addEventListener("click", handleClick);

    return () => {
      element.removeEventListener("click", handleClick);
    };
  }, []);

  return (
    <button ref={ref} className="s-btn s-btn__icon s-btn__facebook bar-md">
      <svg
        aria-hidden="true"
        class="svg-icon iconFacebook"
        width="18"
        height="18"
        viewBox="0 0 18 18"
      >
        <path
          d="M3 1a2 2 0 0 0-2 2v12c0 1.1.9 2 2 2h12a2 2 0 0 0 2-2V3a2 2 0 0 0-2-2H3Zm6.55 16v-6.2H7.46V8.4h2.09V6.61c0-2.07 1.26-3.2 3.1-3.2.88 0 1.64.07 1.87.1v2.16h-1.29c-1 0-1.19.48-1.19 1.18V8.4h2.39l-.31 2.42h-2.08V17h-2.5Z"
          fill="#4167B2"
        ></path>
      </svg>
      &nbsp;Log in with Facebook
    </button>
  );
};

export default FacebookLogin;
