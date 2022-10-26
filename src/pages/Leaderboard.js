import React, { useEffect, useState } from "react";
import "bootstrap/dist/css/bootstrap.min.css";
import Spinner from "react-bootstrap/Spinner";

//components
import LeaderboardCard from "../components/LeaderboardCard";


// Import Dummy Data DELETE LATER
const dummyStandingData = require("./standingdummydata.json");

// const getStandingData = async (token) => {

//   try {
//     return fetch("/standings", {
//       method: "GET",
//       headers: {
//         Authorization: `Bearer ${token}`,
//         "Content-Type": "application/json",
//         "Access-Control-Allow-Origin": "*",
//         "Access-Control-Allow-Headers":
//           "Origin, X-Requested-With, Content-Type, Accept",
//       },
//     }).then((res) => res.json());
//   } catch (err) {}
// };

// const getToken = async () => {
//   return fetch("/user/login", {
//     method: "POST",
//     body: JSON.stringify({
//       email: "patrickreynoldscoding@gmail.com",
//       password: "password",
//     }),
//     headers: {
//       "Content-Type": "application/json",
//       "Access-Control-Allow-Origin": "*",
//       "Access-Control-Allow-Headers":
//         "Origin, X-Requested-With, Content-Type, Accept",
//     },
//   }).then((res) => res.json());
// };

export default function Leaderboard() {
  const [standingData, setStandingData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // getToken()
    //   .then((res) => getStandingData(res.data.token))
    //   .then((res) => setStandingData(JSON.parse(JSON.stringify(res.data))))
    //   .then(() => setLoading(false))
    //   .catch(console.error);

    setLoading(false)


  }, []);





  return (
    <div>
      {loading ? (
        <Spinner animation="border" variant="primary" />
      ) : (
        dummyStandingData.data.map((group) => { // UPDATE THIS TO standingData and uncomment the above useEffect
          return (
            <div key={group._id}> 
              {group.teams.map((team) => {
                return <LeaderboardCard team={team} key={team.team_id}/>
              })}

            </div>
          );
        })
      )}
    </div>
  );
}
