export const getUserData = async () => {
  try {
    const res = await fetch("http://localhost:8080/private_data", {
      //change endpoint as needed
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Headers":
          "Origin, X-Requested-With, Content-Type, Accept",
      },
    })
      .then((res) => res.json())
      .then((data) => {
        console.log(data.userInfo.attrs.team_name);
      });
  } catch (err) {}
};
