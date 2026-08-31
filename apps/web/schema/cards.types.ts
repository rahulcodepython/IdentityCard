import { z } from "zod";

export const resendResponseSchema = z.object({ message: z.string() });
export type ResendResponse = z.infer<typeof resendResponseSchema>;
