import { z } from "zod"

export const MemberResponseSchema = z.object({
  id: z.string().uuid(),
  userId: z.string().uuid(),
  name: z.string(),
  email: z.string().email(),
  roles: z.array(z.string()),
  createdAt: z.string().datetime(),
  status: z.string(),
  assignedEvents: z.array(z.string()),
})

export type MemberResponse = z.infer<typeof MemberResponseSchema>

export const ListMembersResponseSchema = z.array(MemberResponseSchema)
