import { apiRequest } from "@/react-query/client"
import { z } from "zod"
import { ListMembersResponseSchema, type MemberResponse } from "@/schema/members.types"

export async function getMembers(): Promise<MemberResponse[]> {
  return apiRequest(
    {
      method: "GET",
      url: "/organizations/members",
    },
    ListMembersResponseSchema
  )
}

export async function deleteMember(memberId: string): Promise<void> {
  return apiRequest(
    {
      method: "DELETE",
      url: `/organizations/members/${memberId}`,
    },
    z.any()
  )
}
