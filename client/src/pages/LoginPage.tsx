import { useLocation } from "react-router";
import AuthCard from "../components/layout/AuthCard";
import LoginForm from "../components/auth/LoginForm";
import Alert from "../components/ui/alert/Alert";

export default function LoginPage() {
  const location = useLocation();
  const justSignedUp = location.state?.justSignedUp;

  return (
    <AuthCard
      title="Welcome back"
      subtitle="Log in to continue to your account."
    >
      {justSignedUp && (
        <Alert variant="success" className="mb-5">
          Account created. You can log in now.
        </Alert>
      )}
      <LoginForm />
    </AuthCard>
  );
}
