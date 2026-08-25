import z from "zod";

export function ValidateInput<T>(
  input: T,
  schema: z.ZodObject<any>,
): Record<string, string> {
  const result = schema.safeParse(input);
  if (result.success) {
    return {};
  } else {
    const { fieldErrors } = z.flattenError(result.error);

    // Convert arrays of messages into single strings
    const errors: Record<string, string> = {};
    for (const key in fieldErrors) {
      if (fieldErrors[key] && fieldErrors[key].length > 0) {
        errors[key] = fieldErrors[key].join(", ");
      }
    }

    return errors;
  }
}
