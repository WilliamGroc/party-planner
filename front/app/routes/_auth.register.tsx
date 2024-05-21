import { ActionFunction } from "@remix-run/node";
import { Form } from "@remix-run/react";
import { css } from "styled-system/css";

export const action: ActionFunction = async ({ request }) => {
  return new Response(null, {
    status: 200,
    headers: {
      "Set-Cookie": `auth=123; Path=/; HttpOnly; SameSite=Strict;`,
    },
  });
}

export default function AuthRegister() {
  return (
    <Form className={css({
      display: 'flex',
      flexDir: 'column',
      w: '33%'
    })}>
      <input type="text" name="username" placeholder="Username" />
      <input type="text" name="email" placeholder="Email" />
      <input type="password" name="password" placeholder="Password" />
      <input type="password" name="confirmPassword" placeholder="Confirm password" />
      <button type="submit">Login</button>
    </Form>
  );
}