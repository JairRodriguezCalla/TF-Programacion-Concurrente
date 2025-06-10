import pandas as pd
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report, accuracy_score
from sklearn.preprocessing import LabelEncoder
import joblib

# Cargar el dataset
df = pd.read_csv('./data/facturacion.csv')
df = df.dropna()

# Tomar muestra de 20,000 registros
df = df.sample(n=20000, random_state=42)

# Codificar variables categóricas
label_usos = LabelEncoder()
label_grupo = LabelEncoder()
label_empresa = LabelEncoder()

df['USO_ENC'] = label_usos.fit_transform(df['USO'])
df['GRUPO_ENC'] = label_grupo.fit_transform(df['GRUPO'])
df['EMPRESA_ENC'] = label_empresa.fit_transform(df['COD_EMPRESA'])

# Variables independientes y objetivo
X = df[['PROMEDIO_CONSUMO', 'USO_ENC', 'GRUPO_ENC', 'EMPRESA_ENC']]
y = df['COD_TARIFA']

# Separar entrenamiento y prueba
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Entrenar el modelo
modelo = RandomForestClassifier(n_estimators=100, random_state=42)
modelo.fit(X_train, y_train)

# Evaluar el modelo
y_pred = modelo.predict(X_test)
print("✅ Resultados del modelo:")
print("Accuracy:", accuracy_score(y_test, y_pred))
print("\nReporte de clasificación:\n")
print(classification_report(y_test, y_pred, zero_division=0))

# Guardar el modelo entrenado
joblib.dump(modelo, 'rf_model.pkl')
print("\n✅ Modelo entrenado y guardado como rf_model.pkl")
