import { useEffect } from "react";

// https://hackernoon.com/how-to-add-script-tags-in-react
// <div id="fb-root"></div>
// <script asyhnc defer crossorigin="anonymous" src="https://connect.facebook.net/en_US/sdk.js#xfbml=1&version=v15.0" nonce="QDIDASpD"></script>

export default function useFacebookJDK() {
  useEffect(() => {
    const body = document.querySelector("body");
    const fbDiv = document.createElement("div");
    const script = document.createElement("script", {
      "data-test": "",
      defer: "",
    });

    fbDiv.setAttribute("id", "fb-root");
    script.setAttribute("crossorigin", "anonymous");
    script.setAttribute("nonce", "yvchInmo");
    script.setAttribute(
      "src",
      "https://connect.facebook.net/en_US/sdk.js#xfbml=1&version=v15.0"
    );

    body.appendChild(fbDiv);
    body.appendChild(script);

    return () => {
      body.appendChild(fbDiv);
      body.removeChild(script);
    };
  }, []);
}
