import React, { useEffect, useState } from "react";
import "bootstrap/dist/css/bootstrap.min.css";
import Spinner from "react-bootstrap/Spinner";
import Container from "react-bootstrap/Container";
import "./Matches.css";

const getMatchData = async () => {
  // console.log(`hello ${token}`)

  try {
    return fetch("https://chowie.uk/api/v1/match", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        // "Access-Control-Allow-Origin": "*",
        // "Access-Control-Allow-Headers":
        //   "Origin, X-Requested-With, Content-Type, Accept",
      },
    }).then((res) => res.json());
  } catch (err) {}
};

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

export default function Matches() {
  const [matchData, setMatchData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getMatchData()
      .then((res) => setMatchData(JSON.parse(JSON.stringify(res.data))))
      .then(() => setLoading(false))
      .then(() => console.log(matchData))
      .catch(console.error);
  }, []);

  return (
    <div>
      {loading ? (
        <Spinner animation="border" variant="primary" />
      ) : (
        matchData.map((match) => {
          return (
            <div className="container-fluid" key={match._id}>
              Group: {match.group} - Matchday {match.matchday} of 3<br />
              {match.home_team_en} vs {match.away_team_en} <br />
              {match.local_date}
            </div>
          );
        })
      )}
    </div>
  );
}
