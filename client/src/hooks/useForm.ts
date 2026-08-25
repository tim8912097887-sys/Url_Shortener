import { useCallback, useState } from "react";

type UseFormProps<T extends Record<string, string>> = {
  initialValues: T;
  validate?: (values: T) => Record<keyof T, string>;
  onSubmit: (values: T) => Promise<void>;
};

// A small, dependency-free controlled-form hook. It intentionally mirrors
// the shape of libraries like Formik/react-hook-form (values/errors/touched
// + handleChange/handleBlur/handleSubmit) so it's a drop-in swap later if
// the form gets more complex than this.
export function useForm<T extends Record<string, string>>({
  initialValues,
  validate,
  onSubmit,
}: UseFormProps<T>) {
  const [values, setValues] = useState<T>(initialValues);
  const [errors, setErrors] = useState<Partial<Record<keyof T, string>>>({});
  const [touched, setTouched] = useState<Partial<Record<keyof T, boolean>>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const handleChange: React.ChangeEventHandler<HTMLInputElement> = useCallback(
    (event) => {
      const { name, value, type, checked } = event.target;
      setValues((prev) => ({
        ...prev,
        [name]: type === "checkbox" ? checked : value,
      }));
      setSubmitError(null);
    },
    [],
  );

  const handleBlur: React.FocusEventHandler<HTMLInputElement> = useCallback(
    (event) => {
      const { name } = event.target;
      setTouched((prev) => ({ ...prev, [name]: true }));
      if (validate) {
        setValues((currentValues) => {
          setErrors(validate(currentValues));
          return currentValues;
        });
      }
    },
    [validate],
  );

  const handleSubmit: React.SubmitEventHandler<HTMLFormElement> = useCallback(
    async (event) => {
      event.preventDefault();
      setSubmitError(null);

      const allTouched = Object.keys(values).reduce(
        (acc, key) => ({ ...acc, [key]: true }),
        {},
      );
      setTouched(allTouched);

      const validationErrors = validate ? validate(values) : {};
      setErrors(validationErrors);

      const hasErrors = Object.values(validationErrors).some(Boolean);
      if (hasErrors) return;

      try {
        setIsSubmitting(true);
        await onSubmit(values);
      } catch (err: unknown) {
        if (
          err &&
          typeof err === "object" &&
          "message" in err &&
          typeof err.message === "string"
        ) {
          setSubmitError(err.message);
        } else if (typeof err === "string") {
          setSubmitError(err);
        } else {
          setSubmitError("An unexpected error occurred");
        }
      } finally {
        setIsSubmitting(false);
      }
    },
    [values, validate, onSubmit],
  );

  return {
    values,
    errors,
    touched,
    isSubmitting,
    submitError,
    handleChange,
    handleBlur,
    handleSubmit,
    setValues,
  };
}
