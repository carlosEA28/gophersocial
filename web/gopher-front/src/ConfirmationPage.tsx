import { useNavigate, useParams } from "react-router-dom";
import { API_URL } from "./App";

export const ConfirmationPage = () => {
  const { token = "" } = useParams();
  const redirect = useNavigate();
  const handleConfirma = async () => {
    const res = await fetch(`${API_URL}/users/activate/${token}`, {
      method: "PUT",
    });

    if (res.ok) {
      redirect("/");
    } else {
      alert("Failed");
    }
  };
  return (
    <div>
      <h1>Confirmation</h1>
      <button onClick={handleConfirma}>Click to confirma</button>
    </div>
  );
};
