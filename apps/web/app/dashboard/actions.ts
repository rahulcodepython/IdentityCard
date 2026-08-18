"use server"

import { redirect } from "next/navigation"

import { logout } from "@/lib/api/auth"

export async function logoutAction() {
  await logout()
  redirect("/login")
}
