export type AuthUser = {
  id: string;
  email: string;
  fullName: string;
  role: "admin" | "trainee";
  status: "active" | "disabled";
};

export type LoginInput = {
  email: string;
  password: string;
};

export type LoginResult = {
  user: AuthUser;
  token: string;
};