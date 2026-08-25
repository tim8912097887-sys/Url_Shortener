import z from "zod";

export const loginSchema = z.object({
  email: z.email().optional(),
  password: z
    .string("Password must be a string")
    .min(8, "Password must be at least 8 characters")
    .max(72, "Password must be at most 72 characters")
    .optional(),
});

export type LoginSchemaType = z.infer<typeof loginSchema>;
