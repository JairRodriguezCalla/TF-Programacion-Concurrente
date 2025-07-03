"use client";

import { useState } from "react";

export default function Home() {
  const [consumo, setConsumo] = useState("");
  const [uso, setUso] = useState("");
  const [grupo, setGrupo] = useState("");
  const [empresa, setEmpresa] = useState("");
  const [prediccion, setPrediccion] = useState(null);
  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState(null);

  const handlePredict = async () => {
    // Validación básica de campos
    if (!consumo || !uso || !grupo || !empresa) {
      setError("Por favor, rellena todos los campos para la predicción.");
      setPrediccion(null); // Limpiar predicción anterior si hay campos vacíos
      return;
    }
    setError(null); // Limpiar errores anteriores

    setCargando(true);
    setPrediccion(null);

    try {
      const response = await fetch("http://localhost:8080/predict", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({
          consumo: parseFloat(consumo),
          uso: parseInt(uso),
          grupo: parseInt(grupo),
          empresa: parseInt(empresa),
        }),
      });

      if (!response.ok) {
        let errorMessage = "Error desconocido al predecir la tarifa.";
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
        } catch (jsonError) {
          // Si no se puede parsear como JSON, usar el estado y texto de la respuesta
          errorMessage = `Error ${response.status}: ${response.statusText}`;
        }
        throw new Error(errorMessage);
      }

      const data = await response.json();

      setPrediccion(`Tarifa Predicha: ${data.tarifa}`);
    } catch (err) {
      setError(
        `Error en la predicción: ${err.message}. Asegúrate de que tu API de Go esté corriendo en http://localhost:8080 y que el endpoint /predict exista.`
      );
      console.error("Error de predicción:", err);
    } finally {
      setCargando(false);
    }
  };

  return (
    <main className="container">
      <h1 className="title">Predicción de Categoría Tarifaria</h1>

      <div className="form-group">
        <label htmlFor="consumo" className="form-label">
          Consumo Promedio (kWh):
        </label>
        <input
          type="number"
          id="consumo"
          value={consumo}
          onChange={(e) => setConsumo(e.target.value)}
          className="form-input"
          placeholder="Ej: 350.5"
        />
      </div>

      <div className="form-group">
        <label htmlFor="uso" className="form-label">
          Uso (ej. 1=Residencial, 2=Comercial):
        </label>
        <input
          type="number"
          id="uso"
          value={uso}
          onChange={(e) => setUso(e.target.value)}
          className="form-input"
          placeholder="Ej: 1"
        />
      </div>

      <div className="form-group">
        <label htmlFor="grupo" className="form-label">
          Grupo (ej. 0=Pública, 1=Privada):
        </label>
        <input
          type="number"
          id="grupo"
          value={grupo}
          onChange={(e) => setGrupo(e.target.value)}
          className="form-input"
          placeholder="Ej: 1"
        />
      </div>

      <div className="form-group">
        <label htmlFor="empresa" className="form-label">
          Empresa (número de ID):
        </label>
        <input
          type="number"
          id="empresa"
          value={empresa}
          onChange={(e) => setEmpresa(e.target.value)}
          className="form-input"
          placeholder="Ej: 10"
        />
      </div>

      <button
        onClick={handlePredict}
        disabled={cargando}
        className="predict-button"
      >
        {cargando ? "Prediciendo..." : "Predecir Tarifa"}
      </button>

      {prediccion && (
        <div className="prediction-result">
          <h2 className="prediction-title">Resultado de la Predicción:</h2>
          <p className="prediction-text">{prediccion}</p>
        </div>
      )}

      {error && (
        <div className="error-message">
          <p>{error}</p>
        </div>
      )}
    </main>
  );
}
