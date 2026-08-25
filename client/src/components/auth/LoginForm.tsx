import { Link, useLocation, useNavigate } from "react-router";
import Form from "../ui/form/Form";
import Button from "../ui/button/Button";
import Alert from "../ui/alert/Alert";
import Divider from "../ui/divider/Divider";
import OAuthButtons from "./OAuthButtons";
import { useForm } from "../../hooks/useForm";
import { useAuthStore } from "../../store/useAuthStore";
import { loginSchema, type LoginSchemaType } from "../../schema/login";
import { ValidateInput } from "../../utils/validation";

export default function LoginForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const login = useAuthStore((s) => s.login);
  const redirectTo = location.state?.from?.pathname ?? "/dashboard";

  const {
    values,
    errors,
    touched,
    isSubmitting,
    submitError,
    handleChange,
    handleBlur,
    handleSubmit,
  } = useForm({
    initialValues: { email: "", password: "" },
    validate: (values: LoginSchemaType) => {
      const result = ValidateInput(values, loginSchema);
      return result;
    },
    onSubmit: async (formValues) => {
      await login(formValues);
      navigate(redirectTo, { replace: true });
    },
  });

  return (
    <div className="space-y-6">
      <Form onSubmit={handleSubmit}>
        {submitError && <Alert variant="error">{submitError}</Alert>}

        <Form.Group
          controlId="login-email"
          isInvalid={touched.email && Boolean(errors.email)}
        >
          <Form.Label>Email</Form.Label>
          <Form.Control
            type="email"
            name="email"
            autoComplete="email"
            placeholder="you@example.com"
            value={values.email}
            onChange={handleChange}
            onBlur={handleBlur}
          />
          <Form.Feedback>{touched.email && errors.email}</Form.Feedback>
        </Form.Group>

        <Form.Group
          controlId="login-password"
          isInvalid={touched.password && Boolean(errors.password)}
        >
          <div className="flex items-center justify-between">
            <Form.Label>Password</Form.Label>
            <Link
              to="/forgot-password"
              className="text-xs font-medium text-teal-600 hover:text-teal-700"
            >
              Forgot password?
            </Link>
          </div>
          <Form.Control
            type="password"
            name="password"
            autoComplete="current-password"
            placeholder="••••••••"
            value={values.password}
            onChange={handleChange}
            onBlur={handleBlur}
          />
          <Form.Feedback>{touched.password && errors.password}</Form.Feedback>
        </Form.Group>

        <Button type="submit" fullWidth isLoading={isSubmitting}>
          Log in
        </Button>
      </Form>

      <Divider label="or" />
      <OAuthButtons />

      <p className="text-center text-sm text-slate-500">
        Don&apos;t have an account?{" "}
        <Link
          to="/signup"
          className="font-medium text-teal-600 hover:text-teal-700"
        >
          Sign up
        </Link>
      </p>
    </div>
  );
}
