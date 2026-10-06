import { NavLink, Outlet } from "react-router";

// 全ページで共通の見出しと行き先、各ページの中身は Outlet に入る
export function Layout() {
  return (
    <main>
      <h1>
        <NavLink to="/">Aozora Park</NavLink>
      </h1>

      <nav>
        <NavLink to="/parks">パーク</NavLink>
        <NavLink to="/attractions">アトラクション</NavLink>
        <NavLink to="/ticket-types">券種</NavLink>
        <NavLink to="/inventory">枠</NavLink>
      </nav>

      <Outlet />
    </main>
  );
}
