export const register = async () => {
  try {
    const res = await fetch(`http://localhost:8080/register`, {
      method: "POST",
      body: JSON.stringify({
        email: "patrick@test.com",
        full_name: "paddy reynolds",
        password: "password"
      }),
      headers: {
        "Content-Type": "application/json",
      },
    });
  } catch (err) {}
};

// http://localhost:8080/register

// {
    // "email": "paddy@chowie.uk",
    // "full_name": "Patrick Reynolds",
    // "password": "secret"
// }

        // email: userData.email,
        // full_name: userData.name,
        // password: userData.password