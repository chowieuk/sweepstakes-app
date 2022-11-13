import React from 'react'
import {useRef, useEffect} from 'react';
import {socialLogin, socialLoginErrorHandler} from '../functions/socialLogin';
import useFacebookSDK from "../hooks/useFacebookSDK";

const FacebookLogin = () => {

  useFacebookSDK(); 
  const ref = useRef(null);

  useEffect(() => {
    const handleClick = (e) => {
        e.preventDefault();
        socialLogin("facebook")
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
      <div ref={ref} class="fb-login-button" data-width="" data-size="large" data-button-type="continue_with" data-layout="default" data-auto-logout-link="false" data-use-continue-as="false"></div>
    </div>
  )
}

export default FacebookLogin