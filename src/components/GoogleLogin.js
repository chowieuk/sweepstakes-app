import React from 'react'
import {useRef, useEffect} from 'react';
import {socialLogin, socialLoginErrorHandler} from '../functions/socialLogin';
import normalGoogle from '../images/btn_google_signin_dark_normal_web.png';

const GoogleLogin = () => {

  const ref = useRef(null);

  useEffect(() => {
    const handleClick = (e) => {
        e.preventDefault();
        socialLogin("google")
          .then(() => {
            window.location.replace(window.location.href + "home");
          })
          .catch(socialLoginErrorHandler);
      }

    const element = ref.current;

    element.addEventListener('click', handleClick);

    return () => {
      element.removeEventListener('click', handleClick);
    };
  }, []);

  return (
    <div>
      <img ref={ref} src={normalGoogle} alt="Sign in with Google"/>
    </div>
  )
}

export default GoogleLogin