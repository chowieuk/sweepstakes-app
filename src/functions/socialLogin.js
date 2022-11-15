// social login function

const socialLogin = function socialLogin(prov) {
  return new Promise((resolve, reject) => {
    const url = window.location.href + "?close=true";
    const eurl = encodeURIComponent(url);
    const win = window.open(
      "/auth/" + prov + "/login?id=sweepstakes&from=" + eurl
    );
    const interval = setInterval(() => {
      try {
        if (win.closed) {
          reject(new Error("Login aborted"));
          clearInterval(interval);
          return;
        }
        if (win.location.search.indexOf("error") !== -1) {
          reject(new Error(win.location.search));
          win.close();
          clearInterval(interval);
          return;
        }
        if (win.location.href.indexOf(url) === 0) {
          resolve();
          win.close();
          clearInterval(interval);
        }
      } catch (e) {}
    }, 100);
  });
};

const socialLoginErrorHandler = function socialLoginErrorHandler(err) {
  const status = document.querySelector(".status__label");
  if (err instanceof Response) {
    err.text().then((text) => {
      try {
        const data = JSON.parse(text);
        if (data.error) {
          status.textContent = data.error;
          console.error(data.error);
          return;
        }
      } catch {}
      status.textContent = text;
      console.error(text);
    });
    return;
  }
  status.textContent = err.message;
  console.error(err.message);
};

module.exports = { socialLogin, socialLoginErrorHandler };
