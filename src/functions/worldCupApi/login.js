export const login = async () => {
  try {
    const res = await fetch("/user/login", {
      method: "POST",
      body: JSON.stringify({
        email: "patrickreynoldscoding@gmail.com",
        password: "password",
      }),
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Headers":
          "Origin, X-Requested-With, Content-Type, Accept",
      },
    })
      .then((res) => res.json())
      .then((data) => {
        return data;
      });
  } catch (err) {}
};
