import React from 'react'
import {useRef, useEffect} from 'react';
import {socialLogin, socialLoginErrorHandler} from '../functions/socialLogin';

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
      <button ref={ref}>Login with Google</button>
    </div>
  )
}

export default GoogleLogin